<?php

declare(strict_types=1);

namespace App\Livewire\Dashboard;

use Livewire\Component;

class HumanActions extends Component
{
    public array $pendingQuestions = [];
    public array $reviewGates = [];
    public string $answerText = '';
    public ?string $answeringId = null;

    public function mount(): void
    {
        $this->loadPending();
    }

    public function loadPending(): void
    {
        // Placeholder data — will be replaced with real data from Go backend
        $this->pendingQuestions = [
            [
                'id' => 'q-001',
                'agent' => 'Clotho',
                'job' => '#84',
                'question' => 'Should I apply the fix to both the TCP and Unix socket transports, or just TCP?',
                'asked_at' => now()->subMinutes(8)->toIso8601String(),
                'context' => 'Working on security audit — found unvalidated input in transport layer.',
            ],
        ];

        $this->reviewGates = [
            [
                'id' => 'rg-001',
                'agent' => 'Virgil',
                'job' => '#89',
                'type' => 'pr_review',
                'title' => 'PR #89: fix WebSocket reconnection logic',
                'description' => 'Adds exponential backoff and connection state tracking.',
                'submitted_at' => now()->subMinutes(15)->toIso8601String(),
            ],
        ];
    }

    public function startAnswer(string $questionId): void
    {
        $this->answeringId = $questionId;
        $this->answerText = '';
    }

    public function submitAnswer(): void
    {
        if (! $this->answeringId || trim($this->answerText) === '') {
            return;
        }

        // Remove answered question from list
        $this->pendingQuestions = array_values(
            array_filter($this->pendingQuestions, fn ($q) => $q['id'] !== $this->answeringId)
        );

        $this->answeringId = null;
        $this->answerText = '';
    }

    public function cancelAnswer(): void
    {
        $this->answeringId = null;
        $this->answerText = '';
    }

    public function approveGate(string $gateId): void
    {
        $this->reviewGates = array_values(
            array_filter($this->reviewGates, fn ($g) => $g['id'] !== $gateId)
        );
    }

    public function rejectGate(string $gateId): void
    {
        $this->reviewGates = array_values(
            array_filter($this->reviewGates, fn ($g) => $g['id'] !== $gateId)
        );
    }

    public function render()
    {
        return view('livewire.dashboard.human-actions');
    }
}
