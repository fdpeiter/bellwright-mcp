import type { Metadata } from 'next';
import type { ReactNode } from 'react';
import { IBM_Plex_Mono, IBM_Plex_Sans, Libre_Baskerville } from 'next/font/google';
import './globals.css';

const stewardSerif = Libre_Baskerville({
  weight: ['400', '700'], subsets: ['latin'], variable: '--font-steward-serif',
});
const plexSans = IBM_Plex_Sans({
  weight: ['400', '500', '600'], subsets: ['latin'], variable: '--font-plex-sans',
});
const plexMono = IBM_Plex_Mono({
  weight: ['400', '500', '600'], subsets: ['latin'], variable: '--font-plex-mono',
});

export const metadata: Metadata = {
  title: 'Bellwright Steward',
  description: 'Settlement insights for Bellwright — roster, storage, maps and read-only MCP access',
};

const RootLayout = ({ children }: { children: ReactNode }) => (
  <html lang="en" className="dark">
    <body className={`${stewardSerif.variable} ${plexSans.variable} ${plexMono.variable}`}>
      {children}
    </body>
  </html>
);

export default RootLayout;
