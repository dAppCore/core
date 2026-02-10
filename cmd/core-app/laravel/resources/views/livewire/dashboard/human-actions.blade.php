<div wire:poll.3s="loadPending">
    {{-- Pending questions --}}
    @if (count($pendingQuestions) > 0)
        <div class="mb-6">
            <h3 class="text-sm font-semibold mb-3 flex items-center gap-2">
                <span class="w-2 h-2 rounded-full bg-yellow-500 heartbeat"></span>
                Agent Questions ({{ count($pendingQuestions) }})
            </h3>
            <div class="space-y-3">
                @foreach ($pendingQuestions as $q)
                    <div class="bg-yellow-500/5 border border-yellow-500/30 rounded-lg p-4">
                        <div class="flex items-center gap-2 mb-2">
                            <span class="text-xs font-semibold text-yellow-400">{{ $q['agent'] }}</span>
                            <span class="text-[10px] text-muted font-mono">{{ $q['job'] }}</span>
                            <span class="text-[10px] text-muted">{{ \Carbon\Carbon::parse($q['asked_at'])->diffForHumans(short: true) }}</span>
                        </div>
                        <p class="text-sm text-gray-300 mb-2">{{ $q['question'] }}</p>
                        @if (!empty($q['context']))
                            <p class="text-xs text-muted mb-3">{{ $q['context'] }}</p>
                        @endif

                        @if ($answeringId === $q['id'])
                            <div class="mt-3">
                                <textarea wire:model="answerText"
                                          rows="3"
                                          placeholder="Type your answer..."
                                          class="w-full bg-surface-overlay border border-border rounded-md px-3 py-2 text-sm text-gray-300 placeholder-muted focus:border-accent focus:outline-none resize-none"></textarea>
                                <div class="flex gap-2 mt-2">
                                    <button wire:click="submitAnswer"
                                            class="px-3 py-1.5 text-xs font-medium rounded bg-accent text-surface hover:opacity-90 transition">
                                        Send Answer
                                    </button>
                                    <button wire:click="cancelAnswer"
                                            class="px-3 py-1.5 text-xs font-medium rounded bg-surface-overlay text-muted hover:text-white border border-border transition">
                                        Cancel
                                    </button>
                                </div>
                            </div>
                        @else
                            <button wire:click="startAnswer('{{ $q['id'] }}')"
                                    class="px-3 py-1.5 text-xs font-medium rounded bg-yellow-500/20 text-yellow-400 hover:bg-yellow-500/30 transition">
                                Answer
                            </button>
                        @endif
                    </div>
                @endforeach
            </div>
        </div>
    @endif

    {{-- Review gates --}}
    @if (count($reviewGates) > 0)
        <div>
            <h3 class="text-sm font-semibold mb-3 flex items-center gap-2">
                <span class="w-2 h-2 rounded-full bg-purple-500 heartbeat"></span>
                Review Gates ({{ count($reviewGates) }})
            </h3>
            <div class="space-y-3">
                @foreach ($reviewGates as $gate)
                    <div class="bg-surface-raised border border-purple-500/30 rounded-lg p-4">
                        <div class="flex items-center gap-2 mb-2">
                            <span class="text-xs font-semibold text-purple-400">{{ $gate['agent'] }}</span>
                            <span class="text-[10px] text-muted font-mono">{{ $gate['job'] }}</span>
                            <span class="text-[10px] px-1.5 py-0.5 rounded bg-purple-500/20 text-purple-400 font-medium uppercase">{{ str_replace('_', ' ', $gate['type']) }}</span>
                        </div>
                        <p class="text-sm font-medium text-gray-300 mb-1">{{ $gate['title'] }}</p>
                        <p class="text-xs text-muted mb-3">{{ $gate['description'] }}</p>
                        <div class="flex gap-2">
                            <button wire:click="approveGate('{{ $gate['id'] }}')"
                                    class="px-3 py-1.5 text-xs font-medium rounded bg-green-500/20 text-green-400 hover:bg-green-500/30 transition">
                                Approve
                            </button>
                            <button wire:click="rejectGate('{{ $gate['id'] }}')"
                                    class="px-3 py-1.5 text-xs font-medium rounded bg-red-500/20 text-red-400 hover:bg-red-500/30 transition">
                                Reject
                            </button>
                        </div>
                    </div>
                @endforeach
            </div>
        </div>
    @endif

    @if (count($pendingQuestions) === 0 && count($reviewGates) === 0)
        <div class="text-center py-12 text-muted">
            <svg class="w-8 h-8 mx-auto mb-3 opacity-50" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
            </svg>
            <p class="text-sm">No pending actions. All agents are autonomous.</p>
        </div>
    @endif
</div>
