<?php

declare(strict_types=1);

use App\Models\AgentAllowance;
use App\Models\ModelQuota;
use App\Models\RepoLimit;
use App\Services\AllowanceService;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Route;

/*
|--------------------------------------------------------------------------
| Allowance API Routes
|--------------------------------------------------------------------------
|
| Endpoints for managing agent quotas, checking allowances, and recording
| usage. Protected endpoints use QuotaMiddleware for enforcement.
|
*/

// Health check for quota service
Route::get('/allowances/health', fn () => response()->json(['status' => 'ok']));

// Agent allowance CRUD
Route::prefix('allowances/agents')->group(function () {
    Route::get('/', function () {
        return AgentAllowance::all();
    });

    Route::get('/{agentId}', function (string $agentId) {
        $allowance = AgentAllowance::where('agent_id', $agentId)->first();

        if (! $allowance) {
            return response()->json(['error' => 'not found'], 404);
        }

        return $allowance;
    });

    Route::post('/', function (Request $request) {
        $validated = $request->validate([
            'agent_id' => 'required|string|unique:agent_allowances,agent_id',
            'daily_token_limit' => 'integer|min:0',
            'daily_job_limit' => 'integer|min:0',
            'concurrent_jobs' => 'integer|min:0',
            'max_job_duration_minutes' => 'integer|min:0',
            'model_allowlist' => 'array',
            'model_allowlist.*' => 'string',
        ]);

        return AgentAllowance::create($validated);
    });

    Route::put('/{agentId}', function (Request $request, string $agentId) {
        $allowance = AgentAllowance::where('agent_id', $agentId)->first();

        if (! $allowance) {
            return response()->json(['error' => 'not found'], 404);
        }

        $validated = $request->validate([
            'daily_token_limit' => 'integer|min:0',
            'daily_job_limit' => 'integer|min:0',
            'concurrent_jobs' => 'integer|min:0',
            'max_job_duration_minutes' => 'integer|min:0',
            'model_allowlist' => 'array',
            'model_allowlist.*' => 'string',
        ]);

        $allowance->update($validated);

        return $allowance;
    });

    Route::delete('/{agentId}', function (string $agentId) {
        AgentAllowance::where('agent_id', $agentId)->delete();

        return response()->json(['status' => 'deleted']);
    });
});

// Quota check endpoint
Route::get('/allowances/check/{agentId}', function (Request $request, string $agentId, AllowanceService $svc) {
    $model = $request->query('model', '');

    return response()->json($svc->check($agentId, $model));
});

// Usage reporting endpoint
Route::post('/allowances/usage', function (Request $request, AllowanceService $svc) {
    $validated = $request->validate([
        'agent_id' => 'required|string',
        'job_id' => 'required|string',
        'model' => 'nullable|string',
        'tokens_in' => 'integer|min:0',
        'tokens_out' => 'integer|min:0',
        'event' => 'required|in:job_started,job_completed,job_failed,job_cancelled',
        'timestamp' => 'nullable|date',
    ]);

    $svc->recordUsage($validated);

    return response()->json(['status' => 'recorded']);
});

// Daily reset endpoint
Route::post('/allowances/reset/{agentId}', function (string $agentId, AllowanceService $svc) {
    $svc->resetAgent($agentId);

    return response()->json(['status' => 'reset']);
});

// Model quota management
Route::prefix('allowances/models')->group(function () {
    Route::get('/', fn () => ModelQuota::all());

    Route::post('/', function (Request $request) {
        $validated = $request->validate([
            'model' => 'required|string|unique:model_quotas,model',
            'daily_token_budget' => 'integer|min:0',
            'hourly_rate_limit' => 'integer|min:0',
            'cost_ceiling' => 'integer|min:0',
        ]);

        return ModelQuota::create($validated);
    });

    Route::put('/{model}', function (Request $request, string $model) {
        $quota = ModelQuota::where('model', $model)->first();

        if (! $quota) {
            return response()->json(['error' => 'not found'], 404);
        }

        $validated = $request->validate([
            'daily_token_budget' => 'integer|min:0',
            'hourly_rate_limit' => 'integer|min:0',
            'cost_ceiling' => 'integer|min:0',
        ]);

        $quota->update($validated);

        return $quota;
    });
});
