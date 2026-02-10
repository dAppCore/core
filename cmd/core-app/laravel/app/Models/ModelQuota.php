<?php

declare(strict_types=1);

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class ModelQuota extends Model
{
    protected $fillable = [
        'model',
        'daily_token_budget',
        'hourly_rate_limit',
        'cost_ceiling',
    ];

    protected function casts(): array
    {
        return [
            'daily_token_budget' => 'integer',
            'hourly_rate_limit' => 'integer',
            'cost_ceiling' => 'integer',
        ];
    }
}
