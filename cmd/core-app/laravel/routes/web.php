<?php

declare(strict_types=1);

use Illuminate\Support\Facades\Route;

Route::get('/', function () {
    return view('welcome');
});

// Agentic Dashboard
Route::get('/dashboard', fn () => view('dashboard.index'))->name('dashboard');
Route::get('/dashboard/agents', fn () => view('dashboard.agents'))->name('dashboard.agents');
Route::get('/dashboard/jobs', fn () => view('dashboard.jobs'))->name('dashboard.jobs');
Route::get('/dashboard/activity', fn () => view('dashboard.activity'))->name('dashboard.activity');
