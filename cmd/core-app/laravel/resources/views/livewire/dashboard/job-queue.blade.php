<div wire:poll.5s="loadJobs">
    {{-- Filters --}}
    <div class="flex flex-wrap gap-3 mb-4">
        <select wire:model.live="statusFilter"
                class="bg-surface-overlay border border-border rounded-md px-3 py-1.5 text-xs text-gray-300 focus:border-accent focus:outline-none">
            <option value="all">All statuses</option>
            <option value="queued">Queued</option>
            <option value="in_progress">In Progress</option>
            <option value="review">Review</option>
            <option value="completed">Completed</option>
            <option value="failed">Failed</option>
            <option value="cancelled">Cancelled</option>
        </select>
        <select wire:model.live="agentFilter"
                class="bg-surface-overlay border border-border rounded-md px-3 py-1.5 text-xs text-gray-300 focus:border-accent focus:outline-none">
            <option value="all">All agents</option>
            <option value="Athena">Athena</option>
            <option value="Virgil">Virgil</option>
            <option value="Clotho">Clotho</option>
            <option value="Charon">Charon</option>
        </select>
    </div>

    {{-- Table --}}
    <div class="bg-surface-raised border border-border rounded-lg overflow-hidden">
        <table class="w-full text-sm">
            <thead>
                <tr class="border-b border-border text-xs text-muted uppercase tracking-wider">
                    <th class="text-left px-4 py-3 font-medium">Job</th>
                    <th class="text-left px-4 py-3 font-medium">Issue</th>
                    <th class="text-left px-4 py-3 font-medium">Agent</th>
                    <th class="text-left px-4 py-3 font-medium">Status</th>
                    <th class="text-left px-4 py-3 font-medium">Priority</th>
                    <th class="text-left px-4 py-3 font-medium">Queued</th>
                    <th class="text-right px-4 py-3 font-medium">Actions</th>
                </tr>
            </thead>
            <tbody class="divide-y divide-border">
                @forelse ($this->filteredJobs as $job)
                    <tr class="hover:bg-surface-overlay/50 transition">
                        <td class="px-4 py-3">
                            <div class="font-mono text-xs text-muted">{{ $job['id'] }}</div>
                            <div class="text-xs text-gray-300 mt-0.5 truncate max-w-[200px]">{{ $job['title'] }}</div>
                        </td>
                        <td class="px-4 py-3">
                            <span class="text-accent font-mono text-xs">{{ $job['issue'] }}</span>
                            <div class="text-[11px] text-muted">{{ $job['repo'] }}</div>
                        </td>
                        <td class="px-4 py-3 text-xs">
                            {{ $job['agent'] ?? '—' }}
                        </td>
                        <td class="px-4 py-3">
                            @php
                                $statusColors = [
                                    'queued' => 'bg-yellow-500/20 text-yellow-400',
                                    'in_progress' => 'bg-blue-500/20 text-blue-400',
                                    'review' => 'bg-purple-500/20 text-purple-400',
                                    'completed' => 'bg-green-500/20 text-green-400',
                                    'failed' => 'bg-red-500/20 text-red-400',
                                    'cancelled' => 'bg-gray-500/20 text-gray-400',
                                ];
                            @endphp
                            <span class="text-[10px] px-2 py-0.5 rounded-full font-medium uppercase tracking-wider {{ $statusColors[$job['status']] ?? '' }}">
                                {{ str_replace('_', ' ', $job['status']) }}
                            </span>
                        </td>
                        <td class="px-4 py-3">
                            <span class="text-xs font-mono text-muted">P{{ $job['priority'] }}</span>
                        </td>
                        <td class="px-4 py-3 text-xs text-muted">
                            {{ \Carbon\Carbon::parse($job['queued_at'])->diffForHumans(short: true) }}
                        </td>
                        <td class="px-4 py-3 text-right">
                            <div class="flex items-center justify-end gap-1">
                                @if (in_array($job['status'], ['queued', 'in_progress']))
                                    <button wire:click="cancelJob('{{ $job['id'] }}')"
                                            class="text-[11px] px-2 py-1 rounded bg-red-500/10 text-red-400 hover:bg-red-500/20 transition">
                                        Cancel
                                    </button>
                                @endif
                                @if (in_array($job['status'], ['failed', 'cancelled']))
                                    <button wire:click="retryJob('{{ $job['id'] }}')"
                                            class="text-[11px] px-2 py-1 rounded bg-blue-500/10 text-blue-400 hover:bg-blue-500/20 transition">
                                        Retry
                                    </button>
                                @endif
                            </div>
                        </td>
                    </tr>
                @empty
                    <tr>
                        <td colspan="7" class="px-4 py-8 text-center text-muted text-sm">No jobs match the selected filters.</td>
                    </tr>
                @endforelse
            </tbody>
        </table>
    </div>
</div>
