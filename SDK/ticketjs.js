/**
 * TicketJS — Browser SDK for thermal printer communication via thermalik-bridge
 *
 * Load via CDN:
 *   <script src="https://cdn.jsdelivr.net/gh/saefy/TicketJS@latest/SDK/ticketjs.js"></script>
 *
 * Or from the local bridge:
 *   <script src="http://localhost:8000/thermalik.js"></script>
 */
(function (global) {
  'use strict';

  let _config = {
    apiKey: '',
    apiBaseUrl: 'http://localhost:8000/api/v1',
    bridgeUrl: 'http://localhost:9100',
    printer: null,
  };

  /**
   * Configure the SDK. Call once before printing.
   *
   * @param {object} options
   * @param {string} options.apiKey         - Thermalik API key (trk_...)
   * @param {string} [options.apiBaseUrl]   - Base URL of the Thermalik API (default: http://localhost:8000/api/v1)
   * @param {string} [options.bridgeUrl]    - URL of the local bridge (default: http://localhost:9100)
   * @param {string} [options.printer]      - Default printer ID from bridge config
   */
  function configure(options) {
    _config = Object.assign({}, _config, options);
  }

  /**
   * Print HTML content to a thermal printer.
   *
   * @param {object}  payload
   * @param {string}  payload.html              - HTML string to render and print
   * @param {number}  [payload.printerWidth=80] - Paper width in mm (58 or 80)
   * @param {string}  [payload.encoding='cp437'] - Character encoding
   * @param {string}  [payload.printer]          - Override the default printer ID
   * @returns {Promise<{job_id: string, status: string}>}
   */
  async function print(payload) {
    const {
      html,
      printerWidth = 80,
      encoding = 'cp437',
      printer = _config.printer,
    } = payload;

    if (!html) throw new Error('[TicketJS] print() requires an "html" field');
    if (!printer) throw new Error('[TicketJS] No printer configured. Pass printer in configure() or print()');

    // 1. Render HTML → ESC/POS bytes via Thermalik backend
    const renderRes = await fetch(_config.apiBaseUrl + '/render', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer ' + _config.apiKey,
      },
      body: JSON.stringify({
        html,
        paper_width_mm: printerWidth,
        encoding,
      }),
    });

    if (!renderRes.ok) {
      const body = await renderRes.text().catch(() => '');
      throw new Error('[TicketJS] Render failed (' + renderRes.status + '): ' + body);
    }

    const { bytes } = await renderRes.json();

    if (!Array.isArray(bytes) || bytes.length === 0) {
      throw new Error('[TicketJS] Render returned empty byte array');
    }

    // 2. Send ESC/POS bytes to the local bridge
    const printRes = await fetch(_config.bridgeUrl + '/print', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ printer, bytes }),
    });

    if (!printRes.ok) {
      const body = await printRes.text().catch(() => '');
      throw new Error('[TicketJS] Bridge print failed (' + printRes.status + '): ' + body);
    }

    return printRes.json();
  }

  /**
   * List printers available in the local bridge.
   *
   * @returns {Promise<Array>}
   */
  async function printers() {
    const res = await fetch(_config.bridgeUrl + '/printers');
    if (!res.ok) throw new Error('[TicketJS] Could not list printers (' + res.status + ')');
    const json = await res.json();
    return json.data || [];
  }

  /**
   * Get the status of a print job.
   *
   * @param {string} jobId
   * @returns {Promise<object>}
   */
  async function jobStatus(jobId) {
    const res = await fetch(_config.bridgeUrl + '/jobs/' + jobId);
    if (!res.ok) throw new Error('[TicketJS] Job not found (' + res.status + ')');
    return res.json();
  }

  global.ticketjs = { configure, print, printers, jobStatus };
})(typeof window !== 'undefined' ? window : global);
