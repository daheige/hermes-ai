#!/usr/bin/env bash
root_dir=$(cd "$(dirname "$0")"; cd ..; pwd)

version=$(cat $root_dir/VERSION)

mkdir -p $root_dir/web/build

while IFS= read -r theme; do
    echo "Building theme: $theme"
    rm -rf $root_dir/web/build/$theme
    cd $root_dir/web/$theme
    npm install
    DISABLE_ESLINT_PLUGIN='true' REACT_APP_VERSION=$version npm run build
    cd $root_dir/web
done < $root_dir/web/THEMES
