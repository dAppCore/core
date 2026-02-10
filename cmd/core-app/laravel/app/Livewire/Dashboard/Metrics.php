<?php

declare(strict_types=1);

namespace App\Livewire\Dashboard;

use Livewire\Component;

class Metrics extends Component
{
    public array $stats = [];
    public array $throughputData = [];
    public array $costBreakdown = [];
    public float $budgetUsed = 0;
    public float $budgetLimit = 0;

    public function mount(): void
    {
        $this->loadMetrics();
    }

    public function loadMetrics(): void
    {
        // Placeholder data — will be replaced with real metrics from Go backend
        $this->stats = [
            'jobs_completed' => 12,
            'prs_merged' => 8,
            'tokens_used' => 1_245_800,
            'cost_today' => 18.42,
            'active_agents' => 3,
            'queue_depth' => 4,
        ];

        $this->budgetUsed = 18.42;
        $this->budgetLimit = 50.00;

        // Hourly throughput for chart
        $this->throughputData = [
            ['hour' => '00:00', 'jobs' => 0, 'tokens' => 0],
            ['hour' => '02:00', 'jobs' => 0, 'tokens' => 0],
            ['hour' => '04:00', 'jobs' => 1, 'tokens' => 45_000],
            ['hour' => '06:00', 'jobs' => 2, 'tokens' => 120_000],
            ['hour' => '08:00', 'jobs' => 3, 'tokens' => 195_000],
            ['hour' => '10:00', 'jobs' => 2, 'tokens' => 280_000],
            ['hour' => '12:00', 'jobs' => 1, 'tokens' => 340_000],
            ['hour' => '14:00', 'jobs' => 3, 'tokens' => 450_000],
        ];

        $this->costBreakdown = [
            ['model' => 'claude-opus-4-6', 'cost' => 12.80, 'tokens' => 856_000],
            ['model' => 'claude-sonnet-4-5', 'cost' => 4.20, 'tokens' => 312_000],
            ['model' => 'claude-haiku-4-5', 'cost' => 1.42, 'tokens' => 77_800],
        ];
    }

    public function render()
    {
        return view('livewire.dashboard.metrics');
    }
}
