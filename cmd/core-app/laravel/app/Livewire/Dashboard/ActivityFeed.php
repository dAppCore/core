<?php

declare(strict_types=1);

namespace App\Livewire\Dashboard;

use Livewire\Component;

class ActivityFeed extends Component
{
    public array $entries = [];
    public string $agentFilter = 'all';
    public string $typeFilter = 'all';
    public bool $showOnlyQuestions = false;

    public function mount(): void
    {
        $this->loadEntries();
    }

    public function loadEntries(): void
    {
        // Placeholder data — will be replaced with real-time WebSocket feed
        $this->entries = [
            [
                'id' => 'act-001',
                'agent' => 'Athena',
                'type' => 'code_write',
                'message' => 'Created AgentFleet Livewire component',
                'job' => '#96',
                'timestamp' => now()->subMinutes(2)->toIso8601String(),
                'is_question' => false,
            ],
            [
                'id' => 'act-002',
                'agent' => 'Athena',
                'type' => 'tool_call',
                'message' => 'Read file: cmd/core-app/laravel/composer.json',
                'job' => '#96',
                'timestamp' => now()->subMinutes(5)->toIso8601String(),
                'is_question' => false,
            ],
            [
                'id' => 'act-003',
                'agent' => 'Clotho',
                'type' => 'question',
                'message' => 'Should I apply the fix to both the TCP and Unix socket transports, or just TCP?',
                'job' => '#84',
                'timestamp' => now()->subMinutes(8)->toIso8601String(),
                'is_question' => true,
            ],
            [
                'id' => 'act-004',
                'agent' => 'Virgil',
                'type' => 'pr_created',
                'message' => 'Opened PR #89: fix WebSocket reconnection logic',
                'job' => '#89',
                'timestamp' => now()->subMinutes(15)->toIso8601String(),
                'is_question' => false,
            ],
            [
                'id' => 'act-005',
                'agent' => 'Virgil',
                'type' => 'test_run',
                'message' => 'All 47 tests passed (0.8s)',
                'job' => '#89',
                'timestamp' => now()->subMinutes(18)->toIso8601String(),
                'is_question' => false,
            ],
            [
                'id' => 'act-006',
                'agent' => 'Athena',
                'type' => 'git_push',
                'message' => 'Pushed branch feat/agentic-dashboard',
                'job' => '#96',
                'timestamp' => now()->subMinutes(22)->toIso8601String(),
                'is_question' => false,
            ],
            [
                'id' => 'act-007',
                'agent' => 'Clotho',
                'type' => 'code_write',
                'message' => 'Added input validation for MCP file_write paths',
                'job' => '#84',
                'timestamp' => now()->subMinutes(30)->toIso8601String(),
                'is_question' => false,
            ],
        ];
    }

    public function getFilteredEntriesProperty(): array
    {
        return array_filter($this->entries, function ($entry) {
            if ($this->showOnlyQuestions && !$entry['is_question']) {
                return false;
            }
            if ($this->agentFilter !== 'all' && $entry['agent'] !== $this->agentFilter) {
                return false;
            }
            if ($this->typeFilter !== 'all' && $entry['type'] !== $this->typeFilter) {
                return false;
            }
            return true;
        });
    }

    public function render()
    {
        return view('livewire.dashboard.activity-feed');
    }
}
