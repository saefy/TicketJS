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

The installers verify the SHA-256 checksum of the download. Then configure your
printer, the web origin that will print and your API key:

```bash
thermalik-bridge setup
```

Get an API key and a license at [ticketjs.com](https://ticketjs.com) (*API Keys* and *Installations*).

## Manage it from the browser

```bash
thermalik-bridge config
```

Opens the local admin page (`http://127.0.0.1:9100/admin/`), already signed in.
Add or edit printers, allow web origins, activate with an API key, print a test
page and see recent jobs. **Save and apply** reloads the Bridge without a restart.
The page only answers on this computer; web pages cannot use it.

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
