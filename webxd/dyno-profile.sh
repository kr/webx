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

if test "$PORT" = 2000
then innerport=2001
else innerport=2000
fi

curl -so /tmp/webxd https://api.webx.io/webxd
chmod +x /tmp/webxd
/tmp/webxd start /tmp/webxd 127.0.0.1:$innerport
rm -f /tmp/webxd
# Webxd will fork itself and return once the child is
# fully up and running, ready to serve requests.
# Then it's safe to remove the binary.

# Webxd sends all incoming HTTP requests to $innerport.
# We want the user code to accept them.
PORT=$innerport
unset innerport
