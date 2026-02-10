<?php

declare(strict_types=1);

namespace App\Services\Forgejo;

use Illuminate\Http\Client\PendingRequest;
use Illuminate\Http\Client\Response;
use Illuminate\Support\Facades\Http;
use RuntimeException;

/**
 * Low-level HTTP client for a single Forgejo instance.
 *
 * Wraps the Laravel HTTP client with token auth, retry, and
 * base-URL scoping so callers never deal with raw HTTP details.
 */
class ForgejoClient
{
    private PendingRequest $http;

    public function __construct(
        private readonly string $baseUrl,
        private readonly string $token,
        int $timeout = 30,
        int $retryTimes = 3,
        int $retrySleep = 500,
    ) {
        if ($this->token === '') {
            throw new RuntimeException("Forgejo API token is required for {$this->baseUrl}");
        }

        $this->http = Http::baseUrl(rtrim($this->baseUrl, '/') . '/api/v1')
            ->withHeaders([
                'Authorization' => "token {$this->token}",
                'Accept'        => 'application/json',
                'Content-Type'  => 'application/json',
            ])
            ->timeout($timeout)
            ->retry($retryTimes, $retrySleep, fn (?\Throwable $e, PendingRequest $req): bool =>
                $e instanceof \Illuminate\Http\Client\ConnectionException
            );
    }

    public function baseUrl(): string
    {
        return $this->baseUrl;
    }

    // ----- Generic verbs -----

    /** @return array<string, mixed> */
    public function get(string $path, array $query = []): array
    {
        return $this->decodeOrFail($this->http->get($path, $query));
    }

    /** @return array<string, mixed> */
    public function post(string $path, array $data = []): array
    {
        return $this->decodeOrFail($this->http->post($path, $data));
    }

    /** @return array<string, mixed> */
    public function patch(string $path, array $data = []): array
    {
        return $this->decodeOrFail($this->http->patch($path, $data));
    }

    /** @return array<string, mixed> */
    public function put(string $path, array $data = []): array
    {
        return $this->decodeOrFail($this->http->put($path, $data));
    }

    public function delete(string $path): void
    {
        $response = $this->http->delete($path);

        if ($response->failed()) {
            throw new RuntimeException(
                "Forgejo DELETE {$path} failed [{$response->status()}]: {$response->body()}"
            );
        }
    }

    /**
     * GET a path and return the raw response body as a string.
     * Useful for endpoints that return non-JSON content (e.g. diffs).
     */
    public function getRaw(string $path, array $query = []): string
    {
        $response = $this->http->get($path, $query);

        if ($response->failed()) {
            throw new RuntimeException(
                "Forgejo GET {$path} failed [{$response->status()}]: {$response->body()}"
            );
        }

        return $response->body();
    }

    /**
     * Paginate through all pages of a list endpoint.
     *
     * @return list<array<string, mixed>>
     */
    public function paginate(string $path, array $query = [], int $limit = 50): array
    {
        $all  = [];
        $page = 1;

        do {
            $response = $this->http->get($path, array_merge($query, [
                'page'  => $page,
                'limit' => $limit,
            ]));

            if ($response->failed()) {
                throw new RuntimeException(
                    "Forgejo GET {$path} page {$page} failed [{$response->status()}]: {$response->body()}"
                );
            }

            $items = $response->json();

            if (!is_array($items) || $items === []) {
                break;
            }

            array_push($all, ...$items);

            // Forgejo returns total count in x-total-count header.
            $total = (int) $response->header('x-total-count');
            $page++;
        } while (count($all) < $total);

        return $all;
    }

    // ----- Internals -----

    /** @return array<string, mixed> */
    private function decodeOrFail(Response $response): array
    {
        if ($response->failed()) {
            throw new RuntimeException(
                "Forgejo API error [{$response->status()}]: {$response->body()}"
            );
        }

        return $response->json() ?? [];
    }
}
