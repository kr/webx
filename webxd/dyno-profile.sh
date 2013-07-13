# Dyno .profile script.
#
# This runs in the user's startup shell process.
# We can edit the environment and export new variables.
# If we return, the dyno will continue booting.
# If we exit, the dyno will crash.

case "$PS" in
web.*) ;;
*    ) return ;;
esac
set -e

curl -so /tmp/webxd http://api.webx.io/webxd
chmod +x /tmp/webxd
/tmp/webxd &
rm -f /tmp/webxd
