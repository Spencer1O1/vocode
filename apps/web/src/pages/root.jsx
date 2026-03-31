import vocodeIconUrl from "@vocode/ui/assets/vocode_icon_white.svg?url";
import { Outlet } from "react-router-dom";

import { SITE } from "../site.js";

const navLink =
  "text-sm font-medium text-neutral-400 transition-colors hover:text-white";

const navBtn =
  "inline-flex items-center justify-center rounded-md bg-[#4f81ff] px-4 py-2 text-sm font-semibold text-white shadow-[0_0_24px_-8px_rgba(79,129,255,0.8)] transition-colors hover:bg-[#3d6fe6]";

function Root() {
  return (
    <div className="min-h-screen bg-[#060606] text-neutral-100 antialiased">
      <a
        href="#main"
        className="absolute left-[-10000px] top-auto z-[100] h-1 w-1 overflow-hidden focus:left-4 focus:top-4 focus:h-auto focus:w-auto focus:overflow-visible focus:rounded-md focus:bg-white focus:px-3 focus:py-2 focus:text-sm focus:text-black"
      >
        Skip to content
      </a>
      <header className="sticky top-0 z-50 border-b border-white/[0.08] bg-[#060606]/80 backdrop-blur-md">
        <div className="mx-auto flex h-16 max-w-6xl items-center justify-between px-5">
          <a
            href="/"
            className="flex items-center gap-2.5 text-white focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[#4f81ff]"
          >
            <img src={vocodeIconUrl} alt="" width={36} height={36} />
            <span className="text-base font-semibold tracking-tight">
              {SITE.name}
            </span>
          </a>
          <nav className="flex items-center gap-6" aria-label="Primary">
            <a className={navLink} href={SITE.docsUrl}>
              Docs
            </a>
            <a className={navLink} href={SITE.githubUrl}>
              GitHub
            </a>
            <a
              className={navBtn}
              href={SITE.marketplaceUrl}
              rel="noopener noreferrer"
            >
              Install extension
            </a>
          </nav>
        </div>
      </header>
      <main id="main">
        <Outlet />
      </main>
      <footer className="border-t border-white/[0.08] bg-[#050505] py-14 text-neutral-500">
        <div className="mx-auto flex max-w-6xl flex-col items-start justify-between gap-8 px-5 sm:flex-row sm:items-center">
          <div className="flex items-center gap-2.5">
            <img src={vocodeIconUrl} alt="" width={28} height={28} />
            <div>
              <div className="text-sm font-semibold text-neutral-300">
                {SITE.name}
              </div>
              <div className="text-xs text-neutral-500">
                {SITE.shortTagline}
              </div>
            </div>
          </div>
          <div className="flex flex-wrap gap-6 text-sm">
            <a className="hover:text-white" href={SITE.docsUrl}>
              Documentation
            </a>
            <a className="hover:text-white" href={SITE.githubUrl}>
              GitHub
            </a>
            <a className="hover:text-white" href={SITE.marketplaceUrl}>
              Marketplace
            </a>
          </div>
          <p className="text-xs text-neutral-600">
            © {new Date().getFullYear()} Vocode
          </p>
        </div>
      </footer>
    </div>
  );
}

export default Root;
