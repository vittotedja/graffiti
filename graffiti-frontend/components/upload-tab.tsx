"use client";

import React from "react";
import {Tabs, TabsContent, TabsList, TabsTrigger} from "@/components/ui/tabs";
import {Button} from "@/components/ui/button";
import {ImageIcon, Upload} from "lucide-react";
import {Input} from "./ui/input";
import {Label} from "@/components/ui/label";
import {Textarea} from "@/components/ui/textarea";

interface UploadTabProps {
	mediaUrl: string;
	setMediaUrl: (value: string) => void;
	caption: string;
	setCaption: (value: string) => void;
	handleFileUpload: (e: React.ChangeEvent<HTMLInputElement>) => void;
	handleMockUpload: () => void;
	handlePreview: () => void;
	isPreviewLoaded: boolean;
	renderIframe: () => React.ReactNode;
	handleReset: () => void;
	handleSubmit: () => void;
	isSubmitting: boolean;
}

const UploadTab: React.FC<UploadTabProps> = ({
	mediaUrl,
	setMediaUrl,
	caption,
	setCaption,
	handleFileUpload,
	handleMockUpload,
	handlePreview,
	isPreviewLoaded,
	renderIframe,
	handleReset,
	handleSubmit,
	isSubmitting,
}) => {
	return (
		<div className="grid gap-6">
			<Tabs defaultValue="upload" className="w-full">
				<TabsList className="grid w-full grid-cols-2">
					<TabsTrigger value="upload">Upload</TabsTrigger>
					<TabsTrigger value="embedlink">Embed Link</TabsTrigger>
				</TabsList>
				<TabsContent value="upload" className="mt-4">
					<div className="border-2 border-dashed border-primary/20 rounded-md p-8 text-center">
						<div className="flex flex-col items-center gap-3">
							<ImageIcon className="h-12 w-12 text-muted-foreground" />
							<div className="text-sm text-muted-foreground">
								Drag and drop an image or click to browse
							</div>
							<div className="flex gap-2">
								<Button variant="outline" size="sm" className="mt-2 relative">
									<Upload className="h-4 w-4 mr-2" />
									Upload Image
									<input
										type="file"
										accept="image/*"
										className="absolute inset-0 w-full h-full opacity-0"
										onChange={handleFileUpload}
									/>
								</Button>
								<Button
									onClick={handleMockUpload}
									variant="default"
									size="sm"
									className="mt-2"
								>
									Use Demo Image
								</Button>
							</div>
						</div>
					</div>
				</TabsContent>
				<TabsContent value="embedlink" className="mt-4">
					<div className="border-2 border-primary/20 rounded-md p-8 text-center bg-black/5">
						<div className="space-y-2">
							<Label htmlFor="url">Media URL</Label>
							<div className="flex gap-2">
								<Input
									id="url"
									placeholder="Paste YouTube, TikTok or other media URL"
									value={mediaUrl}
									onChange={(e) => setMediaUrl(e.target.value)}
								/>
								<Button
									type="button"
									variant="outline"
									onClick={handlePreview}
									disabled={!mediaUrl}
								>
									Preview
								</Button>
							</div>
						</div>
						{isPreviewLoaded && (
							<div className="space-y-2 mt-4">
								<div className="flex items-center justify-between">
									<Label>Preview</Label>
									<Button
										type="button"
										variant="ghost"
										size="icon"
										onClick={handleReset}
										className="h-6 w-6"
									>
										{/* You can replace with an actual X icon */}
										<span>X</span>
									</Button>
								</div>
								<div className="rounded-md overflow-hidden border bg-muted">
									{renderIframe()}
								</div>
								<div className="space-y-2 mt-4">
									<Label htmlFor="caption">Caption</Label>
									<Textarea
										id="caption"
										placeholder="Add a caption to your media..."
										value={caption}
										onChange={(e) => setCaption(e.target.value)}
										rows={3}
									/>
								</div>
								<Button
									className="mt-4"
									onClick={handleSubmit}
									disabled={isSubmitting}
								>
									{isSubmitting ? (
										<div className="flex items-center gap-2">
											<div className="h-4 w-4 rounded-full border-2 border-t-transparent border-white animate-spin"></div>
											Posting...
										</div>
									) : (
										<>Post to Wall</>
									)}
								</Button>
							</div>
						)}
					</div>
				</TabsContent>
			</Tabs>
		</div>
	);
};

export default UploadTab;
