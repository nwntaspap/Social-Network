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
            <span className="author">epapamic, </span>
            <span className="author">geoikonomou, </span>
            <span className="author">danikots, </span>
            <span className="author">ekaramet, </span>
            <span className="author">smichail</span>
          </div>
          <span className="project-rights">- SocialNet © 2026 -</span>
        </div>
      </div>
    </footer>
  );
}
