<?php

declare(strict_types=1);

return [
    /*
    |--------------------------------------------------------------------------
    | Default Forgejo Instance
    |--------------------------------------------------------------------------
    |
    | The instance name to use when no explicit instance is specified.
    |
    */
    'default' => env('FORGEJO_DEFAULT', 'forge'),

    /*
    |--------------------------------------------------------------------------
    | Forgejo Instances
    |--------------------------------------------------------------------------
    |
    | Each entry defines a Forgejo instance the platform can talk to.
    | The service auto-routes by matching the configured URL.
    |
    |   url   — Base URL of the Forgejo instance (no trailing slash)
    |   token — Admin API token for the instance
    |
    */
    'instances' => [
        'forge' => [
            'url'   => env('FORGEJO_FORGE_URL', 'https://forge.lthn.ai'),
            'token' => env('FORGEJO_FORGE_TOKEN', ''),
        ],
        'dev' => [
            'url'   => env('FORGEJO_DEV_URL', 'https://dev.lthn.ai'),
            'token' => env('FORGEJO_DEV_TOKEN', ''),
        ],
        'qa' => [
            'url'   => env('FORGEJO_QA_URL', 'https://qa.lthn.ai'),
            'token' => env('FORGEJO_QA_TOKEN', ''),
        ],
    ],

    /*
    |--------------------------------------------------------------------------
    | HTTP Client Settings
    |--------------------------------------------------------------------------
    */
    'timeout'     => (int) env('FORGEJO_TIMEOUT', 30),
    'retry_times' => (int) env('FORGEJO_RETRY_TIMES', 3),
    'retry_sleep' => (int) env('FORGEJO_RETRY_SLEEP', 500),
];
