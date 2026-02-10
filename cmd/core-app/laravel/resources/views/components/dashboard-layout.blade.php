<!DOCTYPE html>
<html lang="en" class="h-full">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{ $title ?? 'Agentic Dashboard' }} — Core</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script>
        tailwind.config = {
            darkMode: 'class',
            theme: {
                extend: {
                    colors: {
                        surface: { DEFAULT: '#0d1117', raised: '#161b22', overlay: '#21262d' },
                        border: { DEFAULT: '#30363d', subtle: '#21262d' },
                        accent: { DEFAULT: '#39d0d8', dim: '#1b6b6f' },
                        success: '#238636',
                        warning: '#d29922',
                        danger: '#da3633',
                        muted: '#8b949e',
                    },
                },
            },
        }
    </script>
    <script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/apexcharts"></script>
    <style>
        [x-cloak] { display: none !important; }
        @keyframes pulse-dot { 0%, 100% { opacity: 1; } 50% { opacity: .4; } }
        .heartbeat { animation: pulse-dot 2s ease-in-out infinite; }
        .scrollbar-thin::-webkit-scrollbar { width: 6px; }
        .scrollbar-thin::-webkit-scrollbar-track { background: transparent; }
        .scrollbar-thin::-webkit-scrollbar-thumb { background: #30363d; border-radius: 3px; }
    </style>
    @livewireStyles
</head>
<body class="h-full bg-surface text-gray-200 antialiased">
    <div class="flex h-full" x-data="{ sidebarOpen: true }">
        {{-- Sidebar --}}
        <aside class="flex flex-col w-56 border-r border-border bg-surface-raised shrink-0 transition-all"
               :class="sidebarOpen ? 'w-56' : 'w-16'">
            <div class="flex items-center gap-2 px-4 h-14 border-b border-border">
                <svg class="w-6 h-6 text-accent shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"/>
                </svg>
                <span class="font-semibold text-sm tracking-wide" x-show="sidebarOpen" x-cloak>Agentic</span>
            </div>
            <nav class="flex-1 py-2 space-y-0.5 px-2">
                <a href="{{ route('dashboard') }}"
                   class="flex items-center gap-3 px-3 py-2 text-sm rounded-md {{ request()->routeIs('dashboard') ? 'bg-surface-overlay text-white' : 'text-muted hover:bg-surface-overlay hover:text-white' }} transition">
                    <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zm10 0a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zm10 0a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z"/></svg>
                    <span x-show="sidebarOpen">Dashboard</span>
                </a>
                <a href="{{ route('dashboard.agents') }}"
                   class="flex items-center gap-3 px-3 py-2 text-sm rounded-md {{ request()->routeIs('dashboard.agents') ? 'bg-surface-overlay text-white' : 'text-muted hover:bg-surface-overlay hover:text-white' }} transition">
                    <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z"/></svg>
                    <span x-show="sidebarOpen">Agent Fleet</span>
                </a>
                <a href="{{ route('dashboard.jobs') }}"
                   class="flex items-center gap-3 px-3 py-2 text-sm rounded-md {{ request()->routeIs('dashboard.jobs') ? 'bg-surface-overlay text-white' : 'text-muted hover:bg-surface-overlay hover:text-white' }} transition">
                    <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v10a2 2 0 002 2h8a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"/></svg>
                    <span x-show="sidebarOpen">Job Queue</span>
                </a>
                <a href="{{ route('dashboard.activity') }}"
                   class="flex items-center gap-3 px-3 py-2 text-sm rounded-md {{ request()->routeIs('dashboard.activity') ? 'bg-surface-overlay text-white' : 'text-muted hover:bg-surface-overlay hover:text-white' }} transition">
                    <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6"/></svg>
                    <span x-show="sidebarOpen">Activity</span>
                </a>
            </nav>
            <div class="border-t border-border p-2">
                <button @click="sidebarOpen = !sidebarOpen"
                        class="flex items-center justify-center w-full px-3 py-2 text-muted hover:text-white rounded-md hover:bg-surface-overlay transition">
                    <svg class="w-4 h-4 transition-transform" :class="sidebarOpen ? '' : 'rotate-180'" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 19l-7-7 7-7m8 14l-7-7 7-7"/></svg>
                </button>
            </div>
        </aside>

        {{-- Main content --}}
        <main class="flex-1 overflow-auto">
            <header class="sticky top-0 z-10 flex items-center justify-between h-14 px-6 border-b border-border bg-surface/80 backdrop-blur">
                <h1 class="text-sm font-semibold">{{ $title ?? 'Dashboard' }}</h1>
                <div class="flex items-center gap-4">
                    <div class="flex items-center gap-2 text-xs text-muted"
                         x-data="{ connected: true }"
                         x-init="
                            setInterval(() => {
                                connected = navigator.onLine;
                            }, 3000)
                         ">
                        <span class="w-2 h-2 rounded-full heartbeat"
                              :class="connected ? 'bg-green-500' : 'bg-red-500'"></span>
                        <span x-text="connected ? 'Connected' : 'Disconnected'"></span>
                    </div>
                    <span class="text-xs text-muted font-mono">{{ now()->format('H:i') }}</span>
                </div>
            </header>
            <div class="p-6">
                {{ $slot }}
            </div>
        </main>
    </div>
    @livewireScripts
</body>
</html>
