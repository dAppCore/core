<?php

declare(strict_types=1);

namespace App\Http\Middleware;

use App\Services\AllowanceService;
use Closure;
use Illuminate\Http\Request;
use Symfony\Component\HttpFoundation\Response;

class QuotaMiddleware
{
    public function __construct(
        private readonly AllowanceService $allowanceService,
    ) {}

    public function handle(Request $request, Closure $next): Response
    {
        $agentId = $request->header('X-Agent-ID', $request->input('agent_id', ''));
        $model = $request->input('model', '');

        if ($agentId === '') {
            return response()->json([
                'error' => 'agent_id is required',
            ], 400);
        }

        $result = $this->allowanceService->check($agentId, $model);

        if (! $result['allowed']) {
            return response()->json([
                'error' => 'quota_exceeded',
                'status' => $result['status'],
                'reason' => $result['reason'],
                'remaining_tokens' => $result['remaining_tokens'],
                'remaining_jobs' => $result['remaining_jobs'],
            ], 429);
        }

        // Attach quota info to request for downstream use
        $request->merge(['_quota' => $result]);

        return $next($request);
    }
}
