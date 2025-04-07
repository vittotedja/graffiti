"use client";

import type React from "react";
import {useEffect, useRef, useState} from "react";
import {Upload, Lock, Globe, ImageIcon, Loader2, X} from "lucide-react";

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
import {fetchWithAuth} from "@/lib/auth";
import {toast} from "sonner";
import {Wall} from "@/types/wall";
import {getPresignedUrl, uploadToS3} from "@/lib/s3-uploader"; // add this!
import Image from "next/image";

interface CreateWallModalProps {
	isOpen: boolean;
	onClose: () => void;
	sentWallData?: Wall | null;
}

export function CreateWallModal({
	isOpen,
	onClose,
	sentWallData,
}: CreateWallModalProps) {
	const [wallData, setWallData] = useState({
		title: "",
		description: "",
		is_public: true,
	});
	const [wallImage, setWallImage] = useState<File | null>(null);
	const [wallPreview, setWallPreview] = useState<string | null>(null);
	const [loading, setLoading] = useState(false);

	useEffect(() => {
		if (sentWallData) {
			setWallData({
				title: sentWallData.title,
				description: sentWallData.description,
				is_public: sentWallData.is_public,
			});
			console.log("Wall data received:", sentWallData);
		} else
			setWallData({
				title: "",
				description: "",
				is_public: true,
			});
	}, [sentWallData]);

	const wallDataRef = useRef<HTMLInputElement>(null);

	const handleWallBgChange = (e: React.ChangeEvent<HTMLInputElement>) => {
		if (e.target.files && e.target.files[0]) {
			const file = e.target.files[0];
			setWallImage(file);
			setWallPreview(URL.createObjectURL(file));
		}
	};

	const handleSubmit = async (e: React.FormEvent) => {
		e.preventDefault();
		setLoading(true);

		try {
			let wallImageUrl = "";

			if (wallImage) {
				const presignedUrlData = await getPresignedUrl(
					`${wallData.title}-${Date.now()}.png`, // unique name
					wallImage
				);
				await uploadToS3(presignedUrlData.presignedUrl, wallImage);
				wallImageUrl = presignedUrlData.publicUrl;
			}

			if (sentWallData) {
				// Update existing wall
				console.log("Updating wall with data:", sentWallData.id);
				const response = await fetchWithAuth(
					`/api/v1/walls/${sentWallData.id}`,
					{
						method: "PUT",
						headers: {"Content-Type": "application/json"},
						body: JSON.stringify({
							...wallData,
							background_image: wallImageUrl || undefined,
						}),
					}
				);

				if (!response.ok) {
					throw new Error("Failed to update wall");
				}

				toast.success("Wall updated successfully!");
			} else {
				const response = await fetchWithAuth("/api/v2/walls", {
					method: "POST",
					headers: {"Content-Type": "application/json"},
					body: JSON.stringify({
						...wallData,
						background_image: wallImageUrl || undefined,
					}),
				});

				if (!response.ok) {
					throw new Error("Failed to create wall");
				}
				toast.success("Wall created successfully!");
			}

			onClose();
		} catch (error) {
			console.error("Error creating wall:", error);
			toast.error("Failed to create wall. Please try again later.");
		} finally {
			setLoading(false);
		}
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
						{/* Wall Title */}
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

						{/* Wall Description */}
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

						{/* Wall Banner Upload */}
						<div className="grid gap-2">
							<Label>Wall Banner</Label>
							<div className="relative border-2 border-dashed border-primary/20 rounded-md p-6 text-center">
								<div className="flex flex-col items-center gap-2">
									{sentWallData?.background_image ? (
										<div className="relative w-full h-40">
											<Image
												src={sentWallData.background_image}
												alt="Wall preview"
												className="w-full h-full object-cover rounded-md"
												width={400}
												height={200}
											/>
											{/* Remove button */}
											<Button
												type="button"
												variant="destructive"
												size="icon"
												className="absolute top-2 right-2 h-8 w-8"
												onClick={() => {
													setWallImage(null);
													setWallPreview(null);
												}}
											>
												<X className="h-4 w-4" />
											</Button>
										</div>
									) : wallPreview ? (
										<div className="relative w-full h-40">
											<Image
												src={wallPreview}
												alt="Wall preview"
												className="w-full h-full object-cover rounded-md"
												width={400}
												height={200}
											/>
											{/* Remove button */}
											<Button
												type="button"
												variant="destructive"
												size="icon"
												className="absolute top-2 right-2 h-8 w-8"
												onClick={() => {
													setWallImage(null);
													setWallPreview(null);
												}}
											>
												<X className="h-4 w-4" />
											</Button>
										</div>
									) : (
										<>
											<ImageIcon className="h-10 w-10 text-muted-foreground" />
											<div className="text-sm text-muted-foreground">
												Drag and drop an image or click to browse
											</div>
											<Button
												type="button"
												variant="outline"
												size="sm"
												className="mt-2"
												onClick={() => wallDataRef.current?.click()}
											>
												<Upload className="h-4 w-4 mr-2" />
												Upload Image
											</Button>
										</>
									)}
								</div>
								<input
									ref={wallDataRef}
									type="file"
									accept="image/*"
									className="hidden"
									onChange={handleWallBgChange}
								/>
							</div>
						</div>

						{/* Privacy Settings */}
						<div className="grid gap-2">
							<Label>Privacy</Label>
							<RadioGroup
								value={wallData.is_public ? "public" : "private"} // <-- controlled here
								onValueChange={(value) =>
									setWallData({...wallData, is_public: value === "public"})
								}
								className="flex gap-4"
							>
								<div className="flex items-center space-x-2">
									<RadioGroupItem value="public" id="public" />
									<Label
										htmlFor="public"
										className="flex items-center gap-1 cursor-pointer"
									>
										<Globe className="h-4 w-4" /> Public
									</Label>
								</div>
								<div className="flex items-center space-x-2">
									<RadioGroupItem value="private" id="private" />
									<Label
										htmlFor="private"
										className="flex items-center gap-1 cursor-pointer"
									>
										<Lock className="h-4 w-4" /> Private
									</Label>
								</div>
							</RadioGroup>
						</div>
					</div>

					<DialogFooter className="sm:justify-end">
						<Button type="button" variant="outline" onClick={onClose}>
							Cancel
						</Button>
						<Button type="submit" disabled={loading}>
							{loading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
							{sentWallData ? "Update Wall" : "Create Wall"}
						</Button>
					</DialogFooter>
				</form>
			</DialogContent>
		</Dialog>
	);
}
