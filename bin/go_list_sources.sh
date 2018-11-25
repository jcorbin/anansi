#!/bin/bash
set -ueo pipefail

# shellcheck disable=SC2016
go list -f '
{{- $dir := .Dir -}}
{{- range .GoFiles -}}
{{- printf "%s/%s" $dir . }}
{{ end -}}
{{- range .TestGoFiles -}}
{{- printf "%s/%s" $dir . }}
{{ end -}}
' "$@"
