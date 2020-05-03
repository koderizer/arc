#! /bin/sh
# /etc/init.d/vizinit

### BEGIN INIT INFO
# Provides: vizinit
# Default-Start:  2 3 4 5
# Default-Stop: 0 1 6
# Short-Description: start and stop viz
# Description: Viz is a gRPC server that would serve arcli
### END INIT INFO

case "$1" in
  start)
    echo "Starting viz server.."
    viz & 
    ;;
  stop)
    echo "Stopping viz server.."
    echo "Do nothing for now"
    ;;
  *)
    echo "Usage: /etc/init.d/vizinit {start|stop}"
    exit 1
    ;;
esac

exit 0