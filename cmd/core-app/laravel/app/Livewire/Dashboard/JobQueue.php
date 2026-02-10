<?php

declare(strict_types=1);

namespace App\Livewire\Dashboard;

use Livewire\Component;

class JobQueue extends Component
{
    public array $jobs = [];
    public string $statusFilter = 'all';
    public string $agentFilter = 'all';

    public function mount(): void
    {
        $this->loadJobs();
    }

    public function loadJobs(): void
    {
        // Placeholder data — will be replaced with real API calls to Go backend
        $this->jobs = [
            [
                'id' => 'job-001',
                'issue' => '#96',
                'repo' => 'host-uk/core',
                'title' => 'feat(agentic): real-time dashboard',
                'agent' => 'Athena',
                'status' => 'in_progress',
                'priority' => 1,
                'queued_at' => now()->subMinutes(45)->toIso8601String(),
                'started_at' => now()->subMinutes(30)->toIso8601String(),
            ],
            [
                'id' => 'job-002',
                'issue' => '#84',
                'repo' => 'host-uk/core',
                'title' => 'fix: security audit findings',
                'agent' => 'Clotho',
                'status' => 'in_progress',
                'priority' => 2,
                'queued_at' => now()->subHours(2)->toIso8601String(),
                'started_at' => now()->subHours(1)->toIso8601String(),
            ],
            [
                'id' => 'job-003',
                'issue' => '#102',
                'repo' => 'host-uk/core',
                'title' => 'feat: add rate limiting to MCP',
                'agent' => null,
                'status' => 'queued',
                'priority' => 3,
                'queued_at' => now()->subMinutes(10)->toIso8601String(),
                'started_at' => null,
            ],
            [
                'id' => 'job-004',
                'issue' => '#89',
                'repo' => 'host-uk/core',
                'title' => 'fix: WebSocket reconnection',
                'agent' => 'Virgil',
                'status' => 'review',
                'priority' => 2,
                'queued_at' => now()->subHours(4)->toIso8601String(),
                'started_at' => now()->subHours(3)->toIso8601String(),
            ],
            [
                'id' => 'job-005',
                'issue' => '#78',
                'repo' => 'host-uk/core',
                'title' => 'docs: update CLAUDE.md',
                'agent' => 'Virgil',
                'status' => 'completed',
                'priority' => 4,
                'queued_at' => now()->subHours(6)->toIso8601String(),
                'started_at' => now()->subHours(5)->toIso8601String(),
            ],
        ];
    }

    public function updatedStatusFilter(): void
    {
        // Livewire auto-updates the view
    }

    public function cancelJob(string $jobId): void
    {
        $this->jobs = array_map(function ($job) use ($jobId) {
            if ($job['id'] === $jobId && in_array($job['status'], ['queued', 'in_progress'])) {
                $job['status'] = 'cancelled';
            }
            return $job;
        }, $this->jobs);
    }

    public function retryJob(string $jobId): void
    {
        $this->jobs = array_map(function ($job) use ($jobId) {
            if ($job['id'] === $jobId && in_array($job['status'], ['failed', 'cancelled'])) {
                $job['status'] = 'queued';
                $job['agent'] = null;
            }
            return $job;
        }, $this->jobs);
    }

    public function getFilteredJobsProperty(): array
    {
        return array_filter($this->jobs, function ($job) {
            if ($this->statusFilter !== 'all' && $job['status'] !== $this->statusFilter) {
                return false;
            }
            if ($this->agentFilter !== 'all' && ($job['agent'] ?? '') !== $this->agentFilter) {
                return false;
            }
            return true;
        });
    }

    public function render()
    {
        return view('livewire.dashboard.job-queue');
    }
}
