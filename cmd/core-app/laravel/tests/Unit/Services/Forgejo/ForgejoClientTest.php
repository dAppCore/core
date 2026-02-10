<?php

declare(strict_types=1);

namespace Tests\Unit\Services\Forgejo;

use App\Services\Forgejo\ForgejoClient;
use Illuminate\Support\Facades\Http;
use Orchestra\Testbench\TestCase;
use RuntimeException;

class ForgejoClientTest extends TestCase
{
    private const BASE_URL = 'https://forge.test';
    private const TOKEN    = 'test-token-abc123';

    // ---- Construction ----

    public function test_constructor_good(): void
    {
        Http::fake();

        $client = new ForgejoClient(self::BASE_URL, self::TOKEN);

        $this->assertSame(self::BASE_URL, $client->baseUrl());
    }

    public function test_constructor_bad_empty_token(): void
    {
        $this->expectException(RuntimeException::class);
        $this->expectExceptionMessage('API token is required');

        new ForgejoClient(self::BASE_URL, '');
    }

    // ---- GET ----

    public function test_get_good(): void
    {
        Http::fake([
            'forge.test/api/v1/repos/owner/repo' => Http::response(['id' => 1, 'name' => 'repo'], 200),
        ]);

        $client = new ForgejoClient(self::BASE_URL, self::TOKEN, retryTimes: 0);
        $result = $client->get('/repos/owner/repo');

        $this->assertSame(1, $result['id']);
        $this->assertSame('repo', $result['name']);
    }

    public function test_get_bad_server_error(): void
    {
        Http::fake([
            'forge.test/api/v1/repos/owner/repo' => Http::response('Internal Server Error', 500),
        ]);

        $client = new ForgejoClient(self::BASE_URL, self::TOKEN, retryTimes: 0);

        $this->expectException(RuntimeException::class);
        $this->expectExceptionMessage('Forgejo API error [500]');

        $client->get('/repos/owner/repo');
    }

    // ---- POST ----

    public function test_post_good(): void
    {
        Http::fake([
            'forge.test/api/v1/repos/owner/repo/issues' => Http::response(['number' => 42], 201),
        ]);

        $client = new ForgejoClient(self::BASE_URL, self::TOKEN, retryTimes: 0);
        $result = $client->post('/repos/owner/repo/issues', ['title' => 'Bug']);

        $this->assertSame(42, $result['number']);
    }

    // ---- PATCH ----

    public function test_patch_good(): void
    {
        Http::fake([
            'forge.test/api/v1/repos/owner/repo/issues/1' => Http::response(['state' => 'closed'], 200),
        ]);

        $client = new ForgejoClient(self::BASE_URL, self::TOKEN, retryTimes: 0);
        $result = $client->patch('/repos/owner/repo/issues/1', ['state' => 'closed']);

        $this->assertSame('closed', $result['state']);
    }

    // ---- PUT ----

    public function test_put_good(): void
    {
        Http::fake([
            'forge.test/api/v1/teams/5/members/alice' => Http::response([], 204),
        ]);

        $client = new ForgejoClient(self::BASE_URL, self::TOKEN, retryTimes: 0);
        $result = $client->put('/teams/5/members/alice');

        $this->assertIsArray($result);
    }

    // ---- DELETE ----

    public function test_delete_good(): void
    {
        Http::fake([
            'forge.test/api/v1/repos/owner/repo/branches/old' => Http::response('', 204),
        ]);

        $client = new ForgejoClient(self::BASE_URL, self::TOKEN, retryTimes: 0);

        // Should not throw
        $client->delete('/repos/owner/repo/branches/old');
        $this->assertTrue(true);
    }

    public function test_delete_bad_not_found(): void
    {
        Http::fake([
            'forge.test/api/v1/repos/owner/repo/branches/gone' => Http::response('Not Found', 404),
        ]);

        $client = new ForgejoClient(self::BASE_URL, self::TOKEN, retryTimes: 0);

        $this->expectException(RuntimeException::class);
        $this->expectExceptionMessage('failed [404]');

        $client->delete('/repos/owner/repo/branches/gone');
    }

    // ---- getRaw ----

    public function test_getRaw_good(): void
    {
        Http::fake([
            'forge.test/api/v1/repos/owner/repo/pulls/1.diff' => Http::response(
                "diff --git a/file.txt b/file.txt\n",
                200,
                ['Content-Type' => 'text/plain'],
            ),
        ]);

        $client = new ForgejoClient(self::BASE_URL, self::TOKEN, retryTimes: 0);
        $diff   = $client->getRaw('/repos/owner/repo/pulls/1.diff');

        $this->assertStringContainsString('diff --git', $diff);
    }

    // ---- Pagination ----

    public function test_paginate_good(): void
    {
        Http::fake([
            'forge.test/api/v1/orgs/myorg/repos?page=1&limit=2' => Http::response(
                [['id' => 1], ['id' => 2]],
                200,
                ['x-total-count' => '3'],
            ),
            'forge.test/api/v1/orgs/myorg/repos?page=2&limit=2' => Http::response(
                [['id' => 3]],
                200,
                ['x-total-count' => '3'],
            ),
        ]);

        $client = new ForgejoClient(self::BASE_URL, self::TOKEN, retryTimes: 0);
        $repos  = $client->paginate('/orgs/myorg/repos', [], 2);

        $this->assertCount(3, $repos);
        $this->assertSame(1, $repos[0]['id']);
        $this->assertSame(3, $repos[2]['id']);
    }

    public function test_paginate_good_empty(): void
    {
        Http::fake([
            'forge.test/api/v1/orgs/empty/repos?page=1&limit=50' => Http::response([], 200),
        ]);

        $client = new ForgejoClient(self::BASE_URL, self::TOKEN, retryTimes: 0);
        $repos  = $client->paginate('/orgs/empty/repos');

        $this->assertSame([], $repos);
    }

    // ---- Auth header ----

    public function test_auth_header_sent(): void
    {
        Http::fake([
            'forge.test/api/v1/user' => Http::response(['login' => 'bot'], 200),
        ]);

        $client = new ForgejoClient(self::BASE_URL, self::TOKEN, retryTimes: 0);
        $client->get('/user');

        Http::assertSent(function ($request) {
            return $request->hasHeader('Authorization', 'token ' . self::TOKEN);
        });
    }
}
