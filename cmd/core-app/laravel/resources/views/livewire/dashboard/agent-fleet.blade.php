<div wire:poll.5s="loadAgents">
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        @foreach ($agents as $agent)
            <div wire:click="selectAgent('{{ $agent['id'] }}')"
                 class="bg-surface-raised border rounded-lg p-4 cursor-pointer transition hover:border-accent
                        {{ $selectedAgent === $agent['id'] ? 'border-accent' : 'border-border' }}">
                {{-- Header --}}
                <div class="flex items-center justify-between mb-3">
                    <div class="flex items-center gap-2">
                        <span class="w-2.5 h-2.5 rounded-full heartbeat
                            {{ $agent['heartbeat'] === 'green' ? 'bg-green-500' : ($agent['heartbeat'] === 'yellow' ? 'bg-yellow-500' : 'bg-red-500') }}"></span>
                        <span class="font-semibold text-sm">{{ $agent['name'] }}</span>
                    </div>
                    <span class="text-[10px] px-2 py-0.5 rounded-full font-medium uppercase tracking-wider
                        {{ $agent['status'] === 'working' ? 'bg-blue-500/20 text-blue-400' : ($agent['status'] === 'idle' ? 'bg-green-500/20 text-green-400' : 'bg-red-500/20 text-red-400') }}">
                        {{ $agent['status'] }}
                    </span>
                </div>

                {{-- Info --}}
                <div class="space-y-1.5 text-xs text-muted">
                    <div class="flex justify-between">
                        <span>Host</span>
                        <span class="text-gray-300 font-mono">{{ $agent['host'] }}</span>
                    </div>
                    <div class="flex justify-between">
                        <span>Model</span>
                        <span class="text-gray-300 font-mono text-[11px]">{{ $agent['model'] }}</span>
                    </div>
                    <div class="flex justify-between">
                        <span>Uptime</span>
                        <span class="text-gray-300">{{ $agent['uptime'] }}</span>
                    </div>
                    @if ($agent['job'])
                        <div class="flex justify-between">
                            <span>Job</span>
                            <span class="text-accent text-[11px]">{{ $agent['job'] }}</span>
                        </div>
                    @endif
                </div>

                {{-- Expanded detail --}}
                @if ($selectedAgent === $agent['id'])
                    <div class="mt-3 pt-3 border-t border-border space-y-1.5 text-xs text-muted">
                        <div class="flex justify-between">
                            <span>Tokens today</span>
                            <span class="text-gray-300">{{ number_format($agent['tokens_today']) }}</span>
                        </div>
                        <div class="flex justify-between">
                            <span>Jobs completed</span>
                            <span class="text-gray-300">{{ $agent['jobs_completed'] }}</span>
                        </div>
                    </div>
                @endif
            </div>
        @endforeach
    </div>
</div>
