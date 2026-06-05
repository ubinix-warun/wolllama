#!/bin/bash
set -e
echo "=== tidying cli ==="
(cd cli && go mod tidy)
echo "=== tidying api ==="
(cd api && go mod tidy)
echo "=== tidying pkg ==="
(cd pkg && go mod tidy)
echo "=== work sync ==="
go work sync
echo "=== done ==="
