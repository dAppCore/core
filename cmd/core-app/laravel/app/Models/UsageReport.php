<?php

declare(strict_types=1);

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class UsageReport extends Model
{
    protected $fillable = [
        'agent_id',
        'job_id',
        'model',
        'tokens_in',
        'tokens_out',
        'event',
        'reported_at',
    ];

    protected function casts(): array
    {
        return [
            'tokens_in' => 'integer',
            'tokens_out' => 'integer',
            'reported_at' => 'datetime',
        ];
    }
}
