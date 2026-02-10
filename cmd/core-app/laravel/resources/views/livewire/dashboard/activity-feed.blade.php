<div wire:poll.3s="loadEntries">
    {{-- Filters --}}
    <div class="flex flex-wrap items-center gap-3 mb-4">
        <select wire:model.live="agentFilter"
                class="bg-surface-overlay border border-border rounded-md px-3 py-1.5 text-xs text-gray-300 focus:border-accent focus:outline-none">
            <option value="all">All agents</option>
            <option value="Athena">Athena</option>
            <option value="Virgil">Virgil</option>
            <option value="Clotho">Clotho</option>
            <option value="Charon">Charon</option>
        </select>
        <select wire:model.live="typeFilter"
                class="bg-surface-overlay border border-border rounded-md px-3 py-1.5 text-xs text-gray-300 focus:border-accent focus:outline-none">
            <option value="all">All types</option>
            <option value="code_write">Code write</option>
            <option value="tool_call">Tool call</option>
            <option value="test_run">Test run</option>
            <option value="pr_created">PR created</option>
            <option value="git_push">Git push</option>
            <option value="question">Question</option>
        </select>
        <label class="flex items-center gap-2 text-xs text-muted cursor-pointer">
            <input type="checkbox" wire:model.live="showOnlyQuestions"
                   class="rounded border-border bg-surface-overlay text-accent focus:ring-accent">
            Waiting for answer only
        </label>
    </div>

    {{-- Feed --}}
    <div class="space-y-2 max-h-[600px] overflow-y-auto scrollbar-thin">
        @forelse ($this->filteredEntries as $entry)
            <div class="bg-surface-raised border rounded-lg px-4 py-3 transition
                        {{ $entry['is_question'] ? 'border-yellow-500/50 bg-yellow-500/5' : 'border-border' }}">
                <div class="flex items-start gap-3">
                    {{-- Type icon --}}
                    @php
                        $typeIcons = [
                            'code_write' => '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4"/>',
                            'tool_call' => '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>',
                            'test_run' => '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>',
                            'pr_created' => '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4"/>',
                            'git_push' => '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16V4m0 0L3 8m4-4l4 4m6 0v12m0 0l4-4m-4 4l-4-4"/>',
                            'question' => '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01"/>',
                        ];
                        $iconPath = $typeIcons[$entry['type']] ?? $typeIcons['tool_call'];
                        $iconColor = $entry['is_question'] ? 'text-yellow-400' : 'text-muted';
                    @endphp
                    <svg class="w-4 h-4 mt-0.5 shrink-0 {{ $iconColor }}" fill="none" viewBox="0 0 24 24" stroke="currentColor">{!! $iconPath !!}</svg>

                    {{-- Content --}}
                    <div class="flex-1 min-w-0">
                        <div class="flex items-center gap-2 mb-0.5">
                            <span class="text-xs font-semibold text-gray-300">{{ $entry['agent'] }}</span>
                            <span class="text-[10px] text-muted font-mono">{{ $entry['job'] }}</span>
                            @if ($entry['is_question'])
                                <span class="text-[10px] px-1.5 py-0.5 rounded bg-yellow-500/20 text-yellow-400 font-medium">NEEDS ANSWER</span>
                            @endif
                        </div>
                        <p class="text-xs text-gray-400 leading-relaxed">{{ $entry['message'] }}</p>
                    </div>

                    {{-- Timestamp --}}
                    <span class="text-[11px] text-muted shrink-0">
                        {{ \Carbon\Carbon::parse($entry['timestamp'])->diffForHumans(short: true) }}
                    </span>
                </div>
            </div>
        @empty
            <div class="text-center py-8 text-muted text-sm">No activity matching filters.</div>
        @endforelse
    </div>
</div>
