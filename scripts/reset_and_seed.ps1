# DB ni butunlay qayta yaratish va to'liq sxema + seed ma'lumotlar
# Foydalanish: PowerShell'da .\scripts\reset_and_seed.ps1

$env:PGPASSWORD = "AvtoMakon123"
$psql = "C:\Program Files\PostgreSQL\16\bin\psql.exe"
$migrations = "C:\Users\ojavl\Projects\AvtoMakon\avtomakon-backend\migrations"

Write-Output "=== DB qayta yaratish ==="
& $psql -U postgres -h localhost -d postgres -c "DROP DATABASE IF EXISTS avtomakon;"
& $psql -U postgres -h localhost -d postgres -c "CREATE DATABASE avtomakon;"

Write-Output ""
Write-Output "=== Migrationsiyalar ==="
$files = @(
    "000001_init.up.sql",
    "000002_seed_demo.up.sql",
    "000003_seed_businesses.up.sql",
    "000004_seed_market.up.sql",
    "000005_seed_chat.up.sql",
    "000006_test_user.up.sql"
)

foreach ($f in $files) {
    Write-Output "  → $f"
    & $psql -U postgres -h localhost -d avtomakon -f "$migrations\$f" 2>&1 | Select-String -Pattern "ERROR|CREATE|INSERT|ALTER" | Select-Object -First 10
    Write-Output ""
}

Write-Output "=== Tekshirish ==="
& $psql -U postgres -h localhost -d avtomakon -c "SELECT 'users' as table_name, COUNT(*) FROM users UNION ALL SELECT 'posts', COUNT(*) FROM posts UNION ALL SELECT 'businesses', COUNT(*) FROM businesses UNION ALL SELECT 'products', COUNT(*) FROM products UNION ALL SELECT 'conversations', COUNT(*) FROM conversations;"
