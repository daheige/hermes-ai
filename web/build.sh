#!/usr/bin/env bash
set -euo pipefail

root_dir=$(cd "$(dirname "$0")"; cd ..; pwd)

version=$(cat "$root_dir/VERSION")

mkdir -p "$root_dir/web/build"

while IFS= read -r theme; do
    [ -n "$theme" ] || continue
    echo "Building theme: $theme"
    rm -rf "$root_dir/web/build/$theme"
    cd "$root_dir/web/$theme"
    npm install --legacy-peer-deps
    # react-scripts 5 + --legacy-peer-deps 会把 ajv@6 提升到 node_modules 顶层，
    # 而 ajv-keywords@5（schema-utils / terser-webpack-plugin 的依赖）需要 ajv@^8，
    # 导致构建时报错 Cannot find module 'ajv/dist/compile/codegen'。
    # 显式安装 ajv@8 作为直接依赖，npm 会将其提升到顶层，即可修复。
    npm install ajv@^8 --save-dev --legacy-peer-deps
    DISABLE_ESLINT_PLUGIN='true' REACT_APP_VERSION="$version" npm run build
    cd "$root_dir/web"
done < "$root_dir/web/THEMES"
