"use client";
import {Upload, ImageIcon, Camera, Brush} from "lucide-react";
import Link from "next/link";

import {Button} from "@/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";
import {Tabs, TabsContent, TabsList, TabsTrigger} from "@/components/ui/tabs";

interface CreatePostModalProps {
	isOpen: boolean;
	onClose: () => void;
	wallId: string;
}

export function CreatePostModal({
	isOpen,
	onClose,
	wallId,
}: CreatePostModalProps) {
	return (
		<Dialog open={isOpen} onOpenChange={onClose}>
			<DialogContent className="sm:max-w-[500px] bg-background border-2 border-primary/20">
				<DialogHeader>
					<DialogTitle className="text-2xl font-graffiti">
						Add New Post
					</DialogTitle>
					<DialogDescription>
						Share a memory or artwork on this wall
					</DialogDescription>
				</DialogHeader>

				<Tabs defaultValue="upload" className="w-full">
					<TabsList className="grid w-full grid-cols-2">
						<TabsTrigger value="upload">Upload</TabsTrigger>
						<TabsTrigger value="camera">Camera</TabsTrigger>
					</TabsList>

					<TabsContent value="upload" className="mt-4">
						<div className="border-2 border-dashed border-primary/20 rounded-md p-8 text-center">
							<div className="flex flex-col items-center gap-3">
								<ImageIcon className="h-12 w-12 text-muted-foreground" />
								<div className="text-sm text-muted-foreground">
									Drag and drop an image or click to browse
								</div>
								<Button variant="outline" size="sm" className="mt-2">
									<Upload className="h-4 w-4 mr-2" />
									Upload Image
								</Button>
							</div>
						</div>
					</TabsContent>

					<TabsContent value="camera" className="mt-4">
						<div className="border-2 border-primary/20 rounded-md p-8 text-center bg-black/5">
							<div className="flex flex-col items-center gap-3">
								<Camera className="h-12 w-12 text-muted-foreground" />
								<div className="text-sm text-muted-foreground">
									Take a photo with your camera
								</div>
								<Button variant="outline" size="sm" className="mt-2">
									<Camera className="h-4 w-4 mr-2" />
									Open Camera
								</Button>
							</div>
						</div>
					</TabsContent>
				</Tabs>

				<div className="flex flex-col gap-2 mt-2">
					<div className="text-sm text-center">
						Want to get creative with your post?
					</div>
					<Link href={`/create?wallId=${wallId}`} onClick={onClose}>
						<Button className="w-full" variant="default">
							<Brush className="h-4 w-4 mr-2" />
							Open Drawing Tools
						</Button>
					</Link>
				</div>

				<DialogFooter className="mt-4">
					<Button variant="outline" onClick={onClose}>
						Cancel
					</Button>
				</DialogFooter>
			</DialogContent>
		</Dialog>
	);
}
