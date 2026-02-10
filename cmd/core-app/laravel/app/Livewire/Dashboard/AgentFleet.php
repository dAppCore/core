<?php

declare(strict_types=1);

namespace App\Livewire\Dashboard;

use Livewire\Component;

class AgentFleet extends Component
{
    /** @var array<int, array{name: string, host: string, model: string, status: string, job: string, heartbeat: string, uptime: string}> */
    public array $agents = [];

    public ?string $selectedAgent = null;

    public function mount(): void
    {
        $this->loadAgents();
    }

    public function loadAgents(): void
    {
        // Placeholder data — will be replaced with real API calls to Go backend
        $this->agents = [
            [
                'id' => 'athena',
                'name' => 'Athena',
                'host' => 'studio.snider.dev',
                'model' => 'claude-opus-4-6',
                'status' => 'working',
                'job' => '#96 agentic dashboard',
                'heartbeat' => 'green',
                'uptime' => '4h 23m',
                'tokens_today' => 142_580,
                'jobs_completed' => 3,
            ],
            [
                'id' => 'virgil',
                'name' => 'Virgil',
                'host' => 'studio.snider.dev',
                'model' => 'claude-opus-4-6',
                'status' => 'idle',
                'job' => '',
                'heartbeat' => 'green',
                'uptime' => '12h 07m',
                'tokens_today' => 89_230,
                'jobs_completed' => 5,
            ],
            [
                'id' => 'clotho',
                'name' => 'Clotho',
                'host' => 'darwin-au',
                'model' => 'claude-sonnet-4-5',
                'status' => 'working',
                'job' => '#84 security audit',
                'heartbeat' => 'yellow',
                'uptime' => '1h 45m',
                'tokens_today' => 34_100,
                'jobs_completed' => 1,
            ],
            [
                'id' => 'charon',
                'name' => 'Charon',
                'host' => 'linux.snider.dev',
                'model' => 'claude-haiku-4-5',
                'status' => 'unhealthy',
                'job' => '',
                'heartbeat' => 'red',
                'uptime' => '0m',
                'tokens_today' => 0,
                'jobs_completed' => 0,
            ],
        ];
    }

    public function selectAgent(string $agentId): void
    {
        $this->selectedAgent = $this->selectedAgent === $agentId ? null : $agentId;
    }

    public function render()
    {
        return view('livewire.dashboard.agent-fleet');
    }
}
