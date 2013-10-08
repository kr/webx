# Dyno .profile script.
#
# This runs in the user's startup shell process.
# We can edit the environment and export new variables.
# If we return, the dyno will continue booting.
# If we exit, the dyno will crash.

case "$DYNO" in
web.*) mode=web ;;
*    ) mode=mon ;;
esac
set -e

curl -so /tmp/webxd https://webx.herokuapp.com/webxd
chmod +x /tmp/webxd
/tmp/webxd $mode &
