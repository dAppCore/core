# Host UK — Laravel Application Container
# PHP 8.3-FPM with all extensions required by the federated monorepo
#
# Build: docker build -f docker/Dockerfile.app -t host-uk/app:latest ..
# (run from host-uk/ workspace root, not core/)

FROM php:8.3-fpm-alpine AS base

# System dependencies
RUN apk add --no-cache \
    git \
    curl \
    libpng-dev \
    libjpeg-turbo-dev \
    freetype-dev \
    libwebp-dev \
    libzip-dev \
    icu-dev \
    oniguruma-dev \
    libxml2-dev \
    linux-headers \
    $PHPIZE_DEPS

# PHP extensions
RUN docker-php-ext-configure gd \
        --with-freetype \
        --with-jpeg \
        --with-webp \
    && docker-php-ext-install -j$(nproc) \
        bcmath \
        exif \
        gd \
        intl \
        mbstring \
        opcache \
        pcntl \
        pdo_mysql \
        soap \
        xml \
        zip

# Redis extension
RUN pecl install redis && docker-php-ext-enable redis

# Composer
COPY --from=composer:2 /usr/bin/composer /usr/bin/composer

# PHP configuration
RUN mv "$PHP_INI_DIR/php.ini-production" "$PHP_INI_DIR/php.ini"
COPY docker/php/opcache.ini $PHP_INI_DIR/conf.d/opcache.ini
COPY docker/php/php-fpm.conf /usr/local/etc/php-fpm.d/zz-host-uk.conf

# --- Build stage ---
FROM base AS build

WORKDIR /app

# Install dependencies first (cache layer)
COPY composer.json composer.lock ./
RUN composer install \
    --no-dev \
    --no-scripts \
    --no-autoloader \
    --prefer-dist \
    --no-interaction

# Copy application
COPY . .

# Generate autoloader and run post-install
RUN composer dump-autoload --optimize --no-dev \
    && php artisan package:discover --ansi

# Build frontend assets
RUN if [ -f package.json ]; then \
        apk add --no-cache nodejs npm && \
        npm ci --production=false && \
        npm run build && \
        rm -rf node_modules; \
    fi

# --- Production stage ---
FROM base AS production

WORKDIR /app

# Copy built application
COPY --from=build /app /app

# Create storage directories
RUN mkdir -p \
    storage/framework/cache/data \
    storage/framework/sessions \
    storage/framework/views \
    storage/logs \
    bootstrap/cache

# Permissions
RUN chown -R www-data:www-data storage bootstrap/cache

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD php-fpm-healthcheck || exit 1

USER www-data

EXPOSE 9000
