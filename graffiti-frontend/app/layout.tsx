import type {Metadata} from "next";
import {Geist, Geist_Mono} from "next/font/google";
import "./globals.css";
import {ThemeProvider} from "@/components/theme-provider";
import {Navbar} from "@/components/navbar";
import {Toaster} from "@/components/ui/sonner";

const geistSans = Geist({
	variable: "--font-geist-sans",
	subsets: ["latin"],
});

const geistMono = Geist_Mono({
	variable: "--font-geist-mono",
	subsets: ["latin"],
});
export const metadata: Metadata = {
	title: "Graffiti - Digital Graffiti Social Media",
	description:
		"Express yourself through digital graffiti and connect with friends",
};

export default function RootLayout({
	children,
}: Readonly<{
	children: React.ReactNode;
}>) {
	return (
		<html lang="en" suppressHydrationWarning>
			<body className={`${geistSans.variable} ${geistMono.variable} font-sans`}>
				<ThemeProvider defaultTheme="system" storageKey="streetwalls-theme">
					<Navbar />
					{children}
					<Toaster richColors />
				</ThemeProvider>
			</body>
		</html>
	);
}
