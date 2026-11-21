# TicketJS

Print thermal receipts (ESC/POS, 58 and 80 mm) straight from a web page.

Your page sends HTML to the **TicketJS Bridge**, a small service running on the
computer next to the printer. The Bridge renders the ticket and prints it over
USB or the network. Ticket content never goes through TicketJS servers.

This repository is the **public release channel** of the Bridge: binaries,
checksums and installers are published under [Releases](https://github.com/saefy/TicketJS/releases).
The source code is maintained privately.

**Documentation:** https://ticketjs.com/docs

## Install the Bridge

macOS / Linux:

```bash
curl -fsSL https://github.com/saefy/TicketJS/releases/latest/download/install-bridge.sh | bash
```

Windows (PowerShell, *Run as administrator*):

```powershell
iwr https://github.com/saefy/TicketJS/releases/latest/download/install-bridge.ps1 -useb | iex
```

The installers verify the SHA-256 checksum of the download. Then install the
service and configure it in the browser (on Windows, from an Administrator PowerShell):

```bash
thermalik-bridge install
```

It starts the Bridge as a system service and opens its local admin page
(`http://127.0.0.1:9100/admin/`). There, add your printer, allow the web origin
that will print, activate with your API key and print a test page. Open the page
again any time with `thermalik-bridge config`; changes apply without a restart.
The page only answers on this computer; web pages cannot use it.

Get an API key and a license at [ticketjs.com](https://ticketjs.com) (*API Keys* and *Installations*).

**Advanced:** to configure from the terminal instead (no browser, scripted
installs), run the wizard `thermalik-bridge setup`.

## Print from your page

```html
<script src="https://ticketjs.com/ticketjs.js"></script>
<script>
  ticketjs.configure({ printer: 'caja-1' });
  ticketjs.print({ html: '<h2>My Shop</h2><p>Total: $120.00</p>' });
</script>
```

No API key goes in the page: printing is local.

## Update

```bash
thermalik-bridge update
```

The Bridge checks this repository for new releases on startup and only logs a
notice; updates are applied when you run the command (checksum verified).

## Help

- [Getting started](https://ticketjs.com/docs#quick-start)
- [Troubleshooting](https://ticketjs.com/docs#troubleshooting)
- [Security and browser permissions](https://ticketjs.com/docs#security)
