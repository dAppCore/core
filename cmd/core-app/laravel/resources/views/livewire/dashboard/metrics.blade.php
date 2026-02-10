<div wire:poll.10s="loadMetrics">
    {{-- Stat cards --}}
    <div class="grid grid-cols-2 lg:grid-cols-3 xl:grid-cols-6 gap-4 mb-6">
        @php
            $statCards = [
                ['label' => 'Jobs Completed', 'value' => $stats['jobs_completed'], 'icon' => 'M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z', 'color' => 'text-green-400'],
                ['label' => 'PRs Merged', 'value' => $stats['prs_merged'], 'icon' => 'M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4', 'color' => 'text-purple-400'],
                ['label' => 'Tokens Used', 'value' => number_format($stats['tokens_used']), 'icon' => 'M7 8h10M7 12h4m1 8l-4-4H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-3l-4 4z', 'color' => 'text-blue-400'],
                ['label' => 'Cost Today', 'value' => '$' . number_format($stats['cost_today'], 2), 'icon' => 'M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z', 'color' => 'text-yellow-400'],
                ['label' => 'Active Agents', 'value' => $stats['active_agents'], 'icon' => 'M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z', 'color' => 'text-accent'],
                ['label' => 'Queue Depth', 'value' => $stats['queue_depth'], 'icon' => 'M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10', 'color' => 'text-orange-400'],
            ];
        @endphp
        @foreach ($statCards as $card)
            <div class="bg-surface-raised border border-border rounded-lg p-4">
                <div class="flex items-center gap-2 mb-2">
                    <svg class="w-4 h-4 {{ $card['color'] }}" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="{{ $card['icon'] }}"/>
                    </svg>
                    <span class="text-[11px] text-muted uppercase tracking-wider">{{ $card['label'] }}</span>
                </div>
                <div class="text-xl font-bold font-mono {{ $card['color'] }}">{{ $card['value'] }}</div>
            </div>
        @endforeach
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {{-- Budget gauge --}}
        <div class="bg-surface-raised border border-border rounded-lg p-5">
            <h3 class="text-sm font-semibold mb-4">Budget</h3>
            <div class="flex items-end gap-3 mb-3">
                <span class="text-3xl font-bold font-mono text-accent">${{ number_format($budgetUsed, 2) }}</span>
                <span class="text-sm text-muted mb-1">/ ${{ number_format($budgetLimit, 2) }}</span>
            </div>
            @php
                $pct = $budgetLimit > 0 ? min(100, ($budgetUsed / $budgetLimit) * 100) : 0;
                $barColor = $pct > 80 ? 'bg-red-500' : ($pct > 60 ? 'bg-yellow-500' : 'bg-accent');
            @endphp
            <div class="w-full h-3 bg-surface-overlay rounded-full overflow-hidden">
                <div class="{{ $barColor }} h-full rounded-full transition-all duration-500" style="width: {{ $pct }}%"></div>
            </div>
            <div class="text-xs text-muted mt-2">{{ number_format($pct, 0) }}% of daily budget used</div>
        </div>

        {{-- Cost breakdown by model --}}
        <div class="bg-surface-raised border border-border rounded-lg p-5">
            <h3 class="text-sm font-semibold mb-4">Cost by Model</h3>
            <div class="space-y-3">
                @foreach ($costBreakdown as $model)
                    @php
                        $modelPct = $budgetUsed > 0 ? ($model['cost'] / $budgetUsed) * 100 : 0;
                        $modelColors = [
                            'claude-opus-4-6' => 'bg-purple-500',
                            'claude-sonnet-4-5' => 'bg-blue-500',
                            'claude-haiku-4-5' => 'bg-green-500',
                        ];
                        $barCol = $modelColors[$model['model']] ?? 'bg-gray-500';
                    @endphp
                    <div>
                        <div class="flex items-center justify-between text-xs mb-1">
                            <span class="font-mono text-gray-300">{{ $model['model'] }}</span>
                            <span class="text-muted">${{ number_format($model['cost'], 2) }} ({{ number_format($model['tokens']) }} tokens)</span>
                        </div>
                        <div class="w-full h-2 bg-surface-overlay rounded-full overflow-hidden">
                            <div class="{{ $barCol }} h-full rounded-full transition-all duration-500" style="width: {{ $modelPct }}%"></div>
                        </div>
                    </div>
                @endforeach
            </div>
        </div>
    </div>

    {{-- Throughput chart --}}
    <div class="bg-surface-raised border border-border rounded-lg p-5 mt-6"
         x-data="{
            chart: null,
            init() {
                this.chart = new ApexCharts(this.$refs.chart, {
                    chart: {
                        type: 'area',
                        height: 240,
                        background: 'transparent',
                        toolbar: { show: false },
                        zoom: { enabled: false },
                    },
                    theme: { mode: 'dark' },
                    colors: ['#39d0d8', '#8b5cf6'],
                    series: [
                        { name: 'Jobs', data: {{ json_encode(array_column($throughputData, 'jobs')) }} },
                        { name: 'Tokens (k)', data: {{ json_encode(array_map(fn($t) => round($t / 1000, 1), array_column($throughputData, 'tokens'))) }} },
                    ],
                    xaxis: {
                        categories: {{ json_encode(array_column($throughputData, 'hour')) }},
                        labels: { style: { colors: '#8b949e', fontSize: '11px' } },
                    },
                    yaxis: [
                        { labels: { style: { colors: '#39d0d8' } }, title: { text: 'Jobs', style: { color: '#39d0d8' } } },
                        { opposite: true, labels: { style: { colors: '#8b5cf6' } }, title: { text: 'Tokens (k)', style: { color: '#8b5cf6' } } },
                    ],
                    grid: { borderColor: '#21262d', strokeDashArray: 3 },
                    stroke: { curve: 'smooth', width: 2 },
                    fill: { type: 'gradient', gradient: { opacityFrom: 0.3, opacityTo: 0.05 } },
                    dataLabels: { enabled: false },
                    legend: { labels: { colors: '#8b949e' } },
                    tooltip: { theme: 'dark' },
                });
                this.chart.render();
            }
         }">
        <h3 class="text-sm font-semibold mb-4">Throughput</h3>
        <div x-ref="chart"></div>
    </div>
</div>
