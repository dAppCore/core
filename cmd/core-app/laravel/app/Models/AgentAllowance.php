<?php

declare(strict_types=1);

namespace App\Models;

use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\HasMany;

class AgentAllowance extends Model
{
    protected $fillable = [
        'agent_id',
        'daily_token_limit',
        'daily_job_limit',
        'concurrent_jobs',
        'max_job_duration_minutes',
        'model_allowlist',
    ];

    protected function casts(): array
    {
        return [
            'daily_token_limit' => 'integer',
            'daily_job_limit' => 'integer',
            'concurrent_jobs' => 'integer',
            'max_job_duration_minutes' => 'integer',
            'model_allowlist' => 'array',
        ];
    }

    public function usageRecords(): HasMany
    {
        return $this->hasMany(QuotaUsage::class, 'agent_id', 'agent_id');
    }

    public function todayUsage(): ?QuotaUsage
    {
        return $this->usageRecords()
            ->where('period_date', now()->toDateString())
            ->first();
    }
}
