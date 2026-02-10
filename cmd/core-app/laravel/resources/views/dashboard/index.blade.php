<x-dashboard-layout title="Dashboard">
    {{-- Metrics overview at top --}}
    <section class="mb-8">
        <livewire:dashboard.metrics />
    </section>

    <div class="grid grid-cols-1 xl:grid-cols-3 gap-6">
        {{-- Left column: Agent fleet + Human actions --}}
        <div class="xl:col-span-2 space-y-6">
            <section>
                <h2 class="text-sm font-semibold text-muted uppercase tracking-wider mb-3">Agent Fleet</h2>
                <livewire:dashboard.agent-fleet />
            </section>

            <section>
                <h2 class="text-sm font-semibold text-muted uppercase tracking-wider mb-3">Job Queue</h2>
                <livewire:dashboard.job-queue />
            </section>
        </div>

        {{-- Right column: Actions + Activity --}}
        <div class="space-y-6">
            <section>
                <h2 class="text-sm font-semibold text-muted uppercase tracking-wider mb-3">Human Actions</h2>
                <livewire:dashboard.human-actions />
            </section>

            <section>
                <h2 class="text-sm font-semibold text-muted uppercase tracking-wider mb-3">Live Activity</h2>
                <livewire:dashboard.activity-feed />
            </section>
        </div>
    </div>
</x-dashboard-layout>
