"use client";

import type React from "react";

import {useState} from "react";
import {Upload, Lock, Globe, ImageIcon} from "lucide-react";

import {Button} from "@/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";
import {Input} from "@/components/ui/input";
import {Label} from "@/components/ui/label";
import {RadioGroup, RadioGroupItem} from "@/components/ui/radio-group";
import {Textarea} from "@/components/ui/textarea";

interface CreateWallModalProps {
	isOpen: boolean;
	onClose: () => void;
	// eslint-disable-next-line @typescript-eslint/no-explicit-any
	onCreateWall: (wallData: any) => void;
}

export function CreateWallModal({
	isOpen,
	onClose,
	onCreateWall,
}: CreateWallModalProps) {
	const [wallData, setWallData] = useState({
		title: "",
		description: "",
		isPrivate: false,
	});

	const handleSubmit = (e: React.FormEvent) => {
		e.preventDefault();
		onCreateWall(wallData);
		onClose();
	};

	return (
		<Dialog open={isOpen} onOpenChange={onClose}>
			<DialogContent className="sm:max-w-[500px] bg-background border-2 border-primary/20">
				<DialogHeader>
					<DialogTitle className="text-2xl font-graffiti">
						Create New Wall
					</DialogTitle>
					<DialogDescription>
						Create a new wall to organize your posts and memories
					</DialogDescription>
				</DialogHeader>
				<form onSubmit={handleSubmit}>
					<div className="grid gap-4 py-4">
						<div className="grid gap-2">
							<Label htmlFor="title">Wall Title</Label>
							<Input
								id="title"
								placeholder="Enter wall title..."
								value={wallData.title}
								onChange={(e) =>
									setWallData({...wallData, title: e.target.value})
								}
								required
							/>
						</div>

						<div className="grid gap-2">
							<Label htmlFor="description">Description</Label>
							<Textarea
								id="description"
								placeholder="What's this wall about?"
								value={wallData.description}
								onChange={(e) =>
									setWallData({...wallData, description: e.target.value})
								}
								className="resize-none"
								rows={3}
							/>
						</div>

						<div className="grid gap-2">
							<Label>Wall Banner</Label>
							<div className="border-2 border-dashed border-primary/20 rounded-md p-6 text-center">
								<div className="flex flex-col items-center gap-2">
									<ImageIcon className="h-10 w-10 text-muted-foreground" />
									<div className="text-sm text-muted-foreground">
										Drag and drop an image or click to browse
									</div>
									<Button
										type="button"
										variant="outline"
										size="sm"
										className="mt-2"
									>
										<Upload className="h-4 w-4 mr-2" />
										Upload Image
									</Button>
								</div>
							</div>
						</div>

						<div className="grid gap-2">
							<Label>Privacy</Label>
							<RadioGroup
								defaultValue="public"
								onValueChange={(value) =>
									setWallData({...wallData, isPrivate: value === "private"})
								}
								className="flex gap-4"
							>
								<div className="flex items-center space-x-2">
									<RadioGroupItem value="public" id="public" />
									<Label
										htmlFor="public"
										className="flex items-center gap-1 cursor-pointer"
									>
										<Globe className="h-4 w-4" />
										Public
									</Label>
								</div>
								<div className="flex items-center space-x-2">
									<RadioGroupItem value="private" id="private" />
									<Label
										htmlFor="private"
										className="flex items-center gap-1 cursor-pointer"
									>
										<Lock className="h-4 w-4" />
										Private
									</Label>
								</div>
							</RadioGroup>
						</div>
					</div>
					<DialogFooter>
						<Button type="button" variant="outline" onClick={onClose}>
							Cancel
						</Button>
						<Button type="submit">Create Wall</Button>
					</DialogFooter>
				</form>
			</DialogContent>
		</Dialog>
	);
}
