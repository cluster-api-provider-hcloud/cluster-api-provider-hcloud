#!/usr/bin/env bash

CMD_VERIFY=verify
CMD_UPDATE=update

util::ensure_bazel() {
    local cmd=$1
    UTIL_TEST_ACTION=$2
    UTIL_TEST_DESC=$3

    case $cmd in
    $CMD_VERIFY)
      local msg="Verifying"
      local bcmd="test --test_output=streamed"
      ;;
    $CMD_UPDATE)
      local msg="Updating"
      local bcmd="run"
      ;;
    *)
      echo "unknown command '${cmd}'">&2
      ;;
    esac

    if [[ -n "${BUILD_WORKSPACE_DIRECTORY:-}" ]] || [[ -n "${TEST_WORKSPACE:-}" ]]; then # Running inside bazel
        echo "${msg} ${UTIL_TEST_DESC}..." >&2
    elif ! command -v bazel &>/dev/null; then
        echo "Install bazel at https://bazel.build" >&2
        exit 1
    else
        (
        set -o xtrace
        bazel $bcmd //hack:${cmd}-${UTIL_TEST_ACTION}
    )
    exit 0
    fi
}

util::before_job() {
    local cmd=$1
    case $cmd in
    $CMD_VERIFY)
      tmpfiles=$TEST_TMPDIR/files
      (
        mkdir -p "$tmpfiles"
        rm -f bazel-*
        cp -aL "." "$tmpfiles"
        export HOME=$(realpath "$TEST_TMPDIR/home")
      )
      OLD_PWD=$(pwd)
      cd "$tmpfiles"
      ;;
    $CMD_UPDATE)
      cd "$BUILD_WORKSPACE_DIRECTORY"
      ;;
    *)
      echo "unknown command '${cmd}'">&2
      ;;
    esac
}

util::after_job() {
    local cmd=$1
    case $cmd in
    $CMD_VERIFY)
      # Avoid diff -N so we handle empty files correctly
      diff=$(diff -upr \
        -x ".git" \
        -x "bazel-*" \
        -x "_output" \
        "$OLD_PWD" "$tmpfiles" 2>/dev/null || true)

      if [[ -n "${diff}" ]]; then
        echo "${diff}" >&2
        echo >&2
        echo "generated ${UTIL_TEST_DESC}. Please run 'bazel run //hack:update-${UTIL_TEST_ACTION}'" >&2
        exit 1
      fi
      echo "SUCCESS ${UTIL_TEST_DESC} up-to-date"
      ;;
    $CMD_UPDATE)
      ;;
    *)
      echo "unknown command '${cmd}'" >&2
      ;;
    esac
}
