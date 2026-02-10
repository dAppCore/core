<?php

declare(strict_types=1);

namespace App\Services\Forgejo;

use RuntimeException;

/**
 * Business-logic layer for Forgejo operations.
 *
 * Manages multiple Forgejo instances (forge, dev, qa) and provides
 * a unified API for issues, pull requests, repositories, and user
 * management. Mirrors the Go pkg/forge API surface.
 */
class ForgejoService
{
    /** @var array<string, ForgejoClient> */
    private array $clients = [];

    private string $defaultInstance;

    /**
     * @param array<string, array{url: string, token: string}> $instances
     */
    public function __construct(
        array $instances,
        string $defaultInstance = 'forge',
        private readonly int $timeout = 30,
        private readonly int $retryTimes = 3,
        private readonly int $retrySleep = 500,
    ) {
        $this->defaultInstance = $defaultInstance;

        foreach ($instances as $name => $cfg) {
            if (($cfg['token'] ?? '') === '') {
                continue; // skip unconfigured instances
            }

            $this->clients[$name] = new ForgejoClient(
                baseUrl: $cfg['url'],
                token: $cfg['token'],
                timeout: $this->timeout,
                retryTimes: $this->retryTimes,
                retrySleep: $this->retrySleep,
            );
        }
    }

    // ----------------------------------------------------------------
    //  Instance resolution
    // ----------------------------------------------------------------

    public function client(?string $instance = null): ForgejoClient
    {
        $name = $instance ?? $this->defaultInstance;

        return $this->clients[$name]
            ?? throw new RuntimeException("Forgejo instance '{$name}' is not configured or has no token");
    }

    /** @return list<string> */
    public function instances(): array
    {
        return array_keys($this->clients);
    }

    // ----------------------------------------------------------------
    //  Issue Operations
    // ----------------------------------------------------------------

    /** @return array<string, mixed> */
    public function createIssue(
        string $owner,
        string $repo,
        string $title,
        string $body = '',
        array $labels = [],
        string $assignee = '',
        ?string $instance = null,
    ): array {
        $data = ['title' => $title, 'body' => $body];

        if ($labels !== []) {
            $data['labels'] = $labels;
        }
        if ($assignee !== '') {
            $data['assignees'] = [$assignee];
        }

        return $this->client($instance)->post("/repos/{$owner}/{$repo}/issues", $data);
    }

    /** @return array<string, mixed> */
    public function updateIssue(
        string $owner,
        string $repo,
        int $number,
        array $fields,
        ?string $instance = null,
    ): array {
        return $this->client($instance)->patch("/repos/{$owner}/{$repo}/issues/{$number}", $fields);
    }

    public function closeIssue(string $owner, string $repo, int $number, ?string $instance = null): array
    {
        return $this->updateIssue($owner, $repo, $number, ['state' => 'closed'], $instance);
    }

    /** @return array<string, mixed> */
    public function addComment(
        string $owner,
        string $repo,
        int $number,
        string $body,
        ?string $instance = null,
    ): array {
        return $this->client($instance)->post(
            "/repos/{$owner}/{$repo}/issues/{$number}/comments",
            ['body' => $body],
        );
    }

    /**
     * @return list<array<string, mixed>>
     */
    public function listIssues(
        string $owner,
        string $repo,
        string $state = 'open',
        int $page = 1,
        int $limit = 50,
        ?string $instance = null,
    ): array {
        return $this->client($instance)->get("/repos/{$owner}/{$repo}/issues", [
            'state' => $state,
            'type'  => 'issues',
            'page'  => $page,
            'limit' => $limit,
        ]);
    }

    // ----------------------------------------------------------------
    //  Pull Request Operations
    // ----------------------------------------------------------------

    /** @return array<string, mixed> */
    public function createPR(
        string $owner,
        string $repo,
        string $head,
        string $base,
        string $title,
        string $body = '',
        ?string $instance = null,
    ): array {
        return $this->client($instance)->post("/repos/{$owner}/{$repo}/pulls", [
            'head'  => $head,
            'base'  => $base,
            'title' => $title,
            'body'  => $body,
        ]);
    }

    public function mergePR(
        string $owner,
        string $repo,
        int $number,
        string $strategy = 'merge',
        ?string $instance = null,
    ): void {
        $this->client($instance)->post("/repos/{$owner}/{$repo}/pulls/{$number}/merge", [
            'Do'                        => $strategy,
            'delete_branch_after_merge' => true,
        ]);
    }

    /**
     * @return list<array<string, mixed>>
     */
    public function listPRs(
        string $owner,
        string $repo,
        string $state = 'open',
        ?string $instance = null,
    ): array {
        return $this->client($instance)->paginate("/repos/{$owner}/{$repo}/pulls", [
            'state' => $state,
        ]);
    }

    public function getPRDiff(string $owner, string $repo, int $number, ?string $instance = null): string
    {
        return $this->client($instance)->getRaw("/repos/{$owner}/{$repo}/pulls/{$number}.diff");
    }

    // ----------------------------------------------------------------
    //  Repository Operations
    // ----------------------------------------------------------------

    /**
     * @return list<array<string, mixed>>
     */
    public function listRepos(string $org, ?string $instance = null): array
    {
        return $this->client($instance)->paginate("/orgs/{$org}/repos");
    }

    /** @return array<string, mixed> */
    public function getRepo(string $owner, string $name, ?string $instance = null): array
    {
        return $this->client($instance)->get("/repos/{$owner}/{$name}");
    }

    /** @return array<string, mixed> */
    public function createBranch(
        string $owner,
        string $repo,
        string $name,
        string $from = '',
        ?string $instance = null,
    ): array {
        $data = ['new_branch_name' => $name];

        if ($from !== '') {
            $data['old_branch_name'] = $from;
        }

        return $this->client($instance)->post("/repos/{$owner}/{$repo}/branches", $data);
    }

    public function deleteBranch(
        string $owner,
        string $repo,
        string $name,
        ?string $instance = null,
    ): void {
        $this->client($instance)->delete("/repos/{$owner}/{$repo}/branches/{$name}");
    }

    // ----------------------------------------------------------------
    //  User / Token Management
    // ----------------------------------------------------------------

    /** @return array<string, mixed> */
    public function createUser(
        string $username,
        string $email,
        string $password,
        ?string $instance = null,
    ): array {
        return $this->client($instance)->post('/admin/users', [
            'username'             => $username,
            'email'                => $email,
            'password'             => $password,
            'must_change_password' => false,
        ]);
    }

    /** @return array<string, mixed> */
    public function createToken(
        string $username,
        string $name,
        array $scopes = [],
        ?string $instance = null,
    ): array {
        $data = ['name' => $name];

        if ($scopes !== []) {
            $data['scopes'] = $scopes;
        }

        return $this->client($instance)->post("/users/{$username}/tokens", $data);
    }

    public function revokeToken(string $username, int $tokenId, ?string $instance = null): void
    {
        $this->client($instance)->delete("/users/{$username}/tokens/{$tokenId}");
    }

    /** @return array<string, mixed> */
    public function addToOrg(
        string $username,
        string $org,
        int $teamId,
        ?string $instance = null,
    ): array {
        return $this->client($instance)->put("/teams/{$teamId}/members/{$username}");
    }

    // ----------------------------------------------------------------
    //  Org Operations
    // ----------------------------------------------------------------

    /**
     * @return list<array<string, mixed>>
     */
    public function listOrgs(?string $instance = null): array
    {
        return $this->client($instance)->paginate('/user/orgs');
    }
}
