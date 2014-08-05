package lib

func reloadScript(port string) string {
	return `
<script type='text/javascript'>
    (function() {
        var poller;

        function connect() {
            var livereloadWebSocket = new WebSocket("ws://localhost:` + port + `/reload/");
            livereloadWebSocket.onmessage = function(msg) {
                livereloadWebSocket.close();
                window.location.reload(true);
            };
            livereloadWebSocket.onopen = function(x) {
              console.log('[ws] Connection opened', new Date());
              if (poller) {
                clearTimeout(poller);
                poller = null;
              }
            };
            livereloadWebSocket.onclose = function() {
              if (!poller) { console.log('[ws] closing'); }
              poll()
            };
            livereloadWebSocket.onerror = function(err) {
              if (!poller) { console.log('[ws] error', err); }
            };
        }

        function poll() {
            poller = setTimeout(function() {
              connect();
            }, 1000);
        }

        connect()
    })()
</script>
`
}
