<?php

declare(strict_types=1);

namespace App\Services;

use App\Models\AgentAllowance;
use App\Models\ModelQuota;
use App\Models\QuotaUsage;
use App\Models\UsageReport;

class AllowanceService
{
    /**
     * Pre-dispatch check: verify agent has remaining allowance.
     *
     * @return array{allowed: bool, status: string, remaining_tokens: int, remaining_jobs: int, reason: ?string}
     */
    public function check(string $agentId, string $model = ''): array
    {
        $allowance = AgentAllowance::where('agent_id', $agentId)->first();

        if (! $allowance) {
            return [
                'allowed' => false,
                'status' => 'exceeded',
                'remaining_tokens' => 0,
                'remaining_jobs' => 0,
                'reason' => 'no allowance configured for agent',
            ];
        }

        $usage = QuotaUsage::firstOrCreate(
            ['agent_id' => $agentId, 'period_date' => now()->toDateString()],
            ['tokens_used' => 0, 'jobs_started' => 0, 'active_jobs' => 0],
        );

        $result = [
            'allowed' => true,
            'status' => 'ok',
            'remaining_tokens' => -1,
            'remaining_jobs' => -1,
            'reason' => null,
        ];

        // Check model allowlist
        if ($model !== '' && ! empty($allowance->model_allowlist)) {
            if (! in_array($model, $allowance->model_allowlist, true)) {
                return array_merge($result, [
                    'allowed' => false,
                    'status' => 'exceeded',
                    'reason' => "model not in allowlist: {$model}",
                ]);
            }
        }

        // Check daily token limit
        if ($allowance->daily_token_limit > 0) {
            $remaining = $allowance->daily_token_limit - $usage->tokens_used;
            $result['remaining_tokens'] = $remaining;

            if ($remaining <= 0) {
                return array_merge($result, [
                    'allowed' => false,
                    'status' => 'exceeded',
                    'reason' => 'daily token limit exceeded',
                ]);
            }

            $ratio = $usage->tokens_used / $allowance->daily_token_limit;
            if ($ratio >= 0.8) {
                $result['status'] = 'warning';
            }
        }

        // Check daily job limit
        if ($allowance->daily_job_limit > 0) {
            $remaining = $allowance->daily_job_limit - $usage->jobs_started;
            $result['remaining_jobs'] = $remaining;

            if ($remaining <= 0) {
                return array_merge($result, [
                    'allowed' => false,
                    'status' => 'exceeded',
                    'reason' => 'daily job limit exceeded',
                ]);
            }
        }

        // Check concurrent jobs
        if ($allowance->concurrent_jobs > 0 && $usage->active_jobs >= $allowance->concurrent_jobs) {
            return array_merge($result, [
                'allowed' => false,
                'status' => 'exceeded',
                'reason' => 'concurrent job limit reached',
            ]);
        }

        // Check global model quota
        if ($model !== '') {
            $modelQuota = ModelQuota::where('model', $model)->first();

            if ($modelQuota && $modelQuota->daily_token_budget > 0) {
                $modelUsage = UsageReport::where('model', $model)
                    ->whereDate('reported_at', now()->toDateString())
                    ->sum(\DB::raw('tokens_in + tokens_out'));

                if ($modelUsage >= $modelQuota->daily_token_budget) {
                    return array_merge($result, [
                        'allowed' => false,
                        'status' => 'exceeded',
                        'reason' => "global model token budget exceeded for: {$model}",
                    ]);
                }
            }
        }

        return $result;
    }

    /**
     * Record usage from an agent runner report.
     */
    public function recordUsage(array $report): void
    {
        $agentId = $report['agent_id'];
        $totalTokens = ($report['tokens_in'] ?? 0) + ($report['tokens_out'] ?? 0);

        $usage = QuotaUsage::firstOrCreate(
            ['agent_id' => $agentId, 'period_date' => now()->toDateString()],
            ['tokens_used' => 0, 'jobs_started' => 0, 'active_jobs' => 0],
        );

        // Persist the raw report
        UsageReport::create([
            'agent_id' => $report['agent_id'],
            'job_id' => $report['job_id'],
            'model' => $report['model'] ?? null,
            'tokens_in' => $report['tokens_in'] ?? 0,
            'tokens_out' => $report['tokens_out'] ?? 0,
            'event' => $report['event'],
            'reported_at' => $report['timestamp'] ?? now(),
        ]);

        match ($report['event']) {
            'job_started' => $usage->increment('jobs_started') || $usage->increment('active_jobs'),
            'job_completed' => $this->handleCompleted($usage, $totalTokens),
            'job_failed' => $this->handleFailed($usage, $totalTokens),
            'job_cancelled' => $this->handleCancelled($usage, $totalTokens),
            default => null,
        };
    }

    /**
     * Reset daily usage counters for an agent.
     */
    public function resetAgent(string $agentId): void
    {
        QuotaUsage::updateOrCreate(
            ['agent_id' => $agentId, 'period_date' => now()->toDateString()],
            ['tokens_used' => 0, 'jobs_started' => 0, 'active_jobs' => 0],
        );
    }

    private function handleCompleted(QuotaUsage $usage, int $totalTokens): void
    {
        $usage->increment('tokens_used', $totalTokens);
        $usage->decrement('active_jobs');
    }

    private function handleFailed(QuotaUsage $usage, int $totalTokens): void
    {
        $returnAmount = intdiv($totalTokens, 2);
        $usage->increment('tokens_used', $totalTokens - $returnAmount);
        $usage->decrement('active_jobs');
    }

    private function handleCancelled(QuotaUsage $usage, int $totalTokens): void
    {
        $usage->decrement('active_jobs');
        // 100% returned — no token charge
    }
}
