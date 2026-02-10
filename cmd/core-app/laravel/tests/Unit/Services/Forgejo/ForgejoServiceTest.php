<?php

declare(strict_types=1);

namespace Tests\Unit\Services\Forgejo;

use App\Services\Forgejo\ForgejoService;
use Illuminate\Support\Facades\Http;
use Orchestra\Testbench\TestCase;
use RuntimeException;

class ForgejoServiceTest extends TestCase
{
    private const INSTANCES = [
        'forge' => ['url' => 'https://forge.test', 'token' => 'tok-forge'],
        'dev'   => ['url' => 'https://dev.test', 'token' => 'tok-dev'],
    ];

    private function service(): ForgejoService
    {
        return new ForgejoService(
            instances: self::INSTANCES,
            defaultInstance: 'forge',
            timeout: 5,
            retryTimes: 0,
            retrySleep: 0,
        );
    }

    // ---- Instance management ----

    public function test_instances_good(): void
    {
        $svc = $this->service();

        $this->assertSame(['forge', 'dev'], $svc->instances());
    }

    public function test_instances_skips_empty_token(): void
    {
        $svc = new ForgejoService(
            instances: [
                'forge' => ['url' => 'https://forge.test', 'token' => 'tok'],
                'qa'    => ['url' => 'https://qa.test', 'token' => ''],
            ],
        );

        $this->assertSame(['forge'], $svc->instances());
    }

    public function test_client_bad_unknown_instance(): void
    {
        $this->expectException(RuntimeException::class);
        $this->expectExceptionMessage("instance 'nope' is not configured");

        $this->service()->client('nope');
    }

    // ---- Issues ----

    public function test_createIssue_good(): void
    {
        Http::fake([
            'forge.test/api/v1/repos/org/repo/issues' => Http::response([
                'number' => 99,
                'title'  => 'New bug',
            ], 201),
        ]);

        $result = $this->service()->createIssue('org', 'repo', 'New bug', 'Description');

        $this->assertSame(99, $result['number']);

        Http::assertSent(fn ($r) => $r['title'] === 'New bug' && $r['body'] === 'Description');
    }

    public function test_createIssue_good_with_labels_and_assignee(): void
    {
        Http::fake([
            'forge.test/api/v1/repos/org/repo/issues' => Http::response(['number' => 1], 201),
        ]);

        $this->service()->createIssue('org', 'repo', 'Task', assignee: 'alice', labels: [1, 2]);

        Http::assertSent(fn ($r) => $r['assignees'] === ['alice'] && $r['labels'] === [1, 2]);
    }

    public function test_closeIssue_good(): void
    {
        Http::fake([
            'forge.test/api/v1/repos/org/repo/issues/5' => Http::response(['state' => 'closed'], 200),
        ]);

        $result = $this->service()->closeIssue('org', 'repo', 5);

        $this->assertSame('closed', $result['state']);
    }

    public function test_addComment_good(): void
    {
        Http::fake([
            'forge.test/api/v1/repos/org/repo/issues/5/comments' => Http::response(['id' => 100], 201),
        ]);

        $result = $this->service()->addComment('org', 'repo', 5, 'LGTM');

        $this->assertSame(100, $result['id']);
    }

    public function test_listIssues_good(): void
    {
        Http::fake([
            'forge.test/api/v1/repos/org/repo/issues*' => Http::response([
                ['number' => 1],
                ['number' => 2],
            ], 200),
        ]);

        $issues = $this->service()->listIssues('org', 'repo');

        $this->assertCount(2, $issues);
    }

    // ---- Pull Requests ----

    public function test_createPR_good(): void
    {
        Http::fake([
            'forge.test/api/v1/repos/org/repo/pulls' => Http::response([
                'number' => 10,
                'title'  => 'Feature X',
            ], 201),
        ]);

        $result = $this->service()->createPR('org', 'repo', 'feat/x', 'main', 'Feature X');

        $this->assertSame(10, $result['number']);
    }

    public function test_mergePR_good(): void
    {
        Http::fake([
            'forge.test/api/v1/repos/org/repo/pulls/10/merge' => Http::response([], 200),
        ]);

        // Should not throw
        $this->service()->mergePR('org', 'repo', 10, 'squash');
        $this->assertTrue(true);
    }

    public function test_getPRDiff_good(): void
    {
        Http::fake([
            'forge.test/api/v1/repos/org/repo/pulls/10.diff' => Http::response(
                "diff --git a/f.go b/f.go\n+new line\n",
                200,
            ),
        ]);

        $diff = $this->service()->getPRDiff('org', 'repo', 10);

        $this->assertStringContainsString('diff --git', $diff);
    }

    // ---- Repositories ----

    public function test_getRepo_good(): void
    {
        Http::fake([
            'forge.test/api/v1/repos/org/core' => Http::response(['full_name' => 'org/core'], 200),
        ]);

        $result = $this->service()->getRepo('org', 'core');

        $this->assertSame('org/core', $result['full_name']);
    }

    public function test_createBranch_good(): void
    {
        Http::fake([
            'forge.test/api/v1/repos/org/repo/branches' => Http::response(['name' => 'feat/y'], 201),
        ]);

        $result = $this->service()->createBranch('org', 'repo', 'feat/y', 'main');

        $this->assertSame('feat/y', $result['name']);

        Http::assertSent(fn ($r) =>
            $r['new_branch_name'] === 'feat/y' && $r['old_branch_name'] === 'main'
        );
    }

    public function test_deleteBranch_good(): void
    {
        Http::fake([
            'forge.test/api/v1/repos/org/repo/branches/old' => Http::response('', 204),
        ]);

        $this->service()->deleteBranch('org', 'repo', 'old');
        $this->assertTrue(true);
    }

    // ---- User / Token Management ----

    public function test_createUser_good(): void
    {
        Http::fake([
            'forge.test/api/v1/admin/users' => Http::response(['login' => 'bot'], 201),
        ]);

        $result = $this->service()->createUser('bot', 'bot@test.io', 's3cret');

        $this->assertSame('bot', $result['login']);

        Http::assertSent(fn ($r) =>
            $r['username'] === 'bot'
            && $r['must_change_password'] === false
        );
    }

    public function test_createToken_good(): void
    {
        Http::fake([
            'forge.test/api/v1/users/bot/tokens' => Http::response(['sha1' => 'abc123'], 201),
        ]);

        $result = $this->service()->createToken('bot', 'ci-token', ['repo', 'user']);

        $this->assertSame('abc123', $result['sha1']);
    }

    public function test_revokeToken_good(): void
    {
        Http::fake([
            'forge.test/api/v1/users/bot/tokens/42' => Http::response('', 204),
        ]);

        $this->service()->revokeToken('bot', 42);
        $this->assertTrue(true);
    }

    // ---- Multi-instance routing ----

    public function test_explicit_instance_routing(): void
    {
        Http::fake([
            'dev.test/api/v1/repos/org/repo' => Http::response(['full_name' => 'org/repo'], 200),
        ]);

        $result = $this->service()->getRepo('org', 'repo', instance: 'dev');

        $this->assertSame('org/repo', $result['full_name']);

        Http::assertSent(fn ($r) => str_contains($r->url(), 'dev.test'));
    }
}
