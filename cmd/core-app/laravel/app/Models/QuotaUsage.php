<?php

declare(strict_types=1);

namespace App\Models;

use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;

class QuotaUsage extends Model
{
    protected $table = 'quota_usage';

    protected $fillable = [
        'agent_id',
        'tokens_used',
        'jobs_started',
        'active_jobs',
        'period_date',
    ];

    protected function casts(): array
    {
        return [
            'tokens_used' => 'integer',
            'jobs_started' => 'integer',
            'active_jobs' => 'integer',
            'period_date' => 'date',
        ];
    }

    public function allowance(): BelongsTo
    {
        return $this->belongsTo(AgentAllowance::class, 'agent_id', 'agent_id');
    }
}
