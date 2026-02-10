<?php

declare(strict_types=1);

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('agent_allowances', function (Blueprint $table) {
            $table->id();
            $table->string('agent_id')->unique();
            $table->bigInteger('daily_token_limit')->default(0);
            $table->integer('daily_job_limit')->default(0);
            $table->integer('concurrent_jobs')->default(1);
            $table->integer('max_job_duration_minutes')->default(0);
            $table->json('model_allowlist')->nullable();
            $table->timestamps();
        });

        Schema::create('quota_usage', function (Blueprint $table) {
            $table->id();
            $table->string('agent_id')->index();
            $table->bigInteger('tokens_used')->default(0);
            $table->integer('jobs_started')->default(0);
            $table->integer('active_jobs')->default(0);
            $table->date('period_date')->index();
            $table->timestamps();

            $table->unique(['agent_id', 'period_date']);
        });

        Schema::create('model_quotas', function (Blueprint $table) {
            $table->id();
            $table->string('model')->unique();
            $table->bigInteger('daily_token_budget')->default(0);
            $table->integer('hourly_rate_limit')->default(0);
            $table->bigInteger('cost_ceiling')->default(0);
            $table->timestamps();
        });

        Schema::create('usage_reports', function (Blueprint $table) {
            $table->id();
            $table->string('agent_id')->index();
            $table->string('job_id')->index();
            $table->string('model')->nullable();
            $table->bigInteger('tokens_in')->default(0);
            $table->bigInteger('tokens_out')->default(0);
            $table->string('event');
            $table->timestamp('reported_at');
            $table->timestamps();
        });

        Schema::create('repo_limits', function (Blueprint $table) {
            $table->id();
            $table->string('repo')->unique();
            $table->integer('max_daily_prs')->default(0);
            $table->integer('max_daily_issues')->default(0);
            $table->integer('cooldown_after_failure_minutes')->default(0);
            $table->timestamps();
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('repo_limits');
        Schema::dropIfExists('usage_reports');
        Schema::dropIfExists('model_quotas');
        Schema::dropIfExists('quota_usage');
        Schema::dropIfExists('agent_allowances');
    }
};
