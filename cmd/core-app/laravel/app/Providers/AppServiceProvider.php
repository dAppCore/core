<?php

declare(strict_types=1);

namespace App\Providers;

use Illuminate\Support\Facades\Artisan;
use Illuminate\Support\ServiceProvider;
use Throwable;

class AppServiceProvider extends ServiceProvider
{
    public function boot(): void
    {
        // Auto-migrate on first boot. Single-user desktop app with
        // SQLite — safe to run on every startup. The --force flag
        // is required in production, --no-interaction prevents prompts.
        try {
            Artisan::call('migrate', [
                '--force' => true,
                '--no-interaction' => true,
            ]);
        } catch (Throwable) {
            // Silently skip — DB might not exist yet (e.g. during
            // composer operations or first extraction).
        }
    }
}
