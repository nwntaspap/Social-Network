/**
 * components/Footer.tsx
 *
 * Mirrors renderFooter() from footer.js.
 * Server component — no interactivity needed.
 */

export default function Footer() {
  return (
    <footer>
      <div className="main-container">
        <div className="footer-container">
          <p>Made with dedication and passion by</p>
          <div className="authors">
            <span className="author">geoikonomou,</span>
            <span className="author">epapamic,</span>
            <span className="author">agkiata,</span>
            <span className="author">sos247</span>
          </div>
          <span>- Forum © 2025 -</span>
        </div>
      </div>
    </footer>
  );
}
