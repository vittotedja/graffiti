"use client";

import type React from "react";

import {useState, useRef, useEffect} from "react";
import {
	Upload,
	ImageIcon,
	Brush,
	Eraser,
	Save,
	Undo,
	Redo,
	Check,
	X,
	ChevronLeft,
} from "lucide-react";

import {Button} from "@/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";
import {Tabs, TabsContent, TabsList, TabsTrigger} from "@/components/ui/tabs";
import {Slider} from "@/components/ui/slider";
import {Label} from "@/components/ui/label";
import {Textarea} from "@/components/ui/textarea";
import {Input} from "./ui/input";
import {toast} from "sonner";

interface EnhancedPostModalProps {
	isOpen: boolean;
	onClose: () => void;
	wallId: string;
	// eslint-disable-next-line @typescript-eslint/no-explicit-any
	onPostCreated: (post: any) => void;
}

export function EnhancedPostModal({
	isOpen,
	onClose,
	// wallId,
	onPostCreated,
}: EnhancedPostModalProps) {
	const [activeTab, setActiveTab] = useState("upload");
	const [selectedImage, setSelectedImage] = useState<string | null>(null);
	const [caption, setCaption] = useState("");
	const [isDrawing, setIsDrawing] = useState(false);
	const [drawingMode, setDrawingMode] = useState("brush");
	const [brushSize, setBrushSize] = useState(5);
	const [brushColor, setBrushColor] = useState("#ff3b30");
	const [drawingHistory, setDrawingHistory] = useState<ImageData[]>([]);
	const [historyIndex, setHistoryIndex] = useState(-1);
	const [isSubmitting, setIsSubmitting] = useState(false);
	const [isPreviewLoaded, setIsPreviewLoaded] = useState(false);

	const [compositeDataUrl, setCompositeDataUrl] = useState<string | null>(null);
	const [mediaUrl, setMediaUrl] = useState<string>("");

	const canvasRef = useRef<HTMLCanvasElement>(null);

	const bgCanvasRef = useRef<HTMLCanvasElement>(null);
	const drawCanvasRef = useRef<HTMLCanvasElement>(null);

	const colors = [
		"#ff3b30", // Red
		"#ff9500", // Orange
		"#ffcc00", // Yellow
		"#34c759", // Green
		"#5ac8fa", // Light Blue
		"#007aff", // Blue
		"#5856d6", // Purple
		"#af52de", // Pink
		"#000000", // Black
		"#ffffff", // White
	];

	useEffect(() => {
		if (selectedImage && bgCanvasRef.current && drawCanvasRef.current) {
			const bgCanvas = bgCanvasRef.current;
			const drawCanvas = drawCanvasRef.current;
			const bgContext = bgCanvas.getContext("2d");
			const drawContext = drawCanvas.getContext("2d");

			if (!bgContext) {
				// Handle the error or return early
				return;
			}

			if (!drawContext) {
				return;
			}

			// Define maximum canvas dimensions
			const MAX_WIDTH = 480;
			const MAX_HEIGHT = 480;

			// Check if we're using a placeholder (empty canvas) scenario
			if (selectedImage.includes("placeholder.svg")) {
				// Instead of loading an image, set a fixed canvas size
				bgCanvas.width = MAX_WIDTH;
				bgCanvas.height = MAX_HEIGHT;
				drawCanvas.width = MAX_WIDTH;
				drawCanvas.height = MAX_HEIGHT;

				// Fill the background canvas with white
				bgContext.fillStyle = "#ffffff";
				bgContext.fillRect(0, 0, MAX_WIDTH, MAX_HEIGHT);

				// Clear the drawing canvas (or you can fill with transparent pixels)
				drawContext.clearRect(0, 0, MAX_WIDTH, MAX_HEIGHT);

				// Initialize drawing history with this blank state
				const initialDrawingState = drawContext.getImageData(
					0,
					0,
					MAX_WIDTH,
					MAX_HEIGHT
				);
				setDrawingHistory([initialDrawingState]);
				setHistoryIndex(0);
			} else {
				// For a real uploaded image, load and scale it if needed
				const img = new Image();
				img.crossOrigin = "anonymous";
				img.onload = () => {
					let width = img.width;
					let height = img.height;
					// Calculate a scale factor so that the image fits within our max dimensions
					const scale = Math.min(MAX_WIDTH / width, MAX_HEIGHT / height, 1); // don't upscale if smaller than max
					width = width * scale;
					height = height * scale;

					// Set both canvases to these dimensions
					bgCanvas.width = width;
					bgCanvas.height = height;
					drawCanvas.width = width;
					drawCanvas.height = height;

					// Draw the image on the background canvas, scaled
					bgContext.drawImage(img, 0, 0, width, height);

					// Initialize drawing canvas with a cleared (transparent) state
					drawContext.clearRect(0, 0, width, height);
					const initialDrawingState = drawContext.getImageData(
						0,
						0,
						width,
						height
					);
					setDrawingHistory([initialDrawingState]);
					setHistoryIndex(0);
				};
				img.src = selectedImage;
			}
		}
	}, [selectedImage, activeTab]);

	// Handle file upload
	const handleFileUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
		const file = e.target.files?.[0];
		if (file) {
			const reader = new FileReader();
			reader.onload = (event) => {
				setSelectedImage(event.target?.result as string);
				setActiveTab("draw");
			};
			reader.readAsDataURL(file);
		}
	};

	// Mock image upload for demo purposes
	const handleMockUpload = () => {
		// Use a placeholder image
		setSelectedImage("/placeholder.svg?height=600&width=600");
		setActiveTab("draw");
	};

	const startDrawing = (e: React.MouseEvent<HTMLCanvasElement>) => {
		if (!drawCanvasRef.current) return;

		const canvas = drawCanvasRef.current;
		const context = canvas.getContext("2d");
		if (!context) return;

		const rect = canvas.getBoundingClientRect();
		const x = (e.clientX - rect.left) * (canvas.width / rect.width);
		const y = (e.clientY - rect.top) * (canvas.height / rect.height);

		context.beginPath();
		context.moveTo(x, y);

		if (drawingMode === "brush") {
			context.globalCompositeOperation = "source-over";
			context.strokeStyle = brushColor;
			context.lineWidth = brushSize;
			context.lineCap = "round";
		} else if (drawingMode === "eraser") {
			// Use destination-out so it only clears drawing on the top layer
			context.globalCompositeOperation = "destination-out";
			context.lineWidth = brushSize;
			context.lineCap = "round";
		}

		setIsDrawing(true);
	};

	const draw = (e: React.MouseEvent<HTMLCanvasElement>) => {
		if (!isDrawing || !drawCanvasRef.current) return;

		const canvas = drawCanvasRef.current;
		const context = canvas.getContext("2d");
		if (!context) return;

		const rect = canvas.getBoundingClientRect();
		const x = (e.clientX - rect.left) * (canvas.width / rect.width);
		const y = (e.clientY - rect.top) * (canvas.height / rect.height);

		context.lineTo(x, y);
		context.stroke();
	};

	const stopDrawing = () => {
		if (!isDrawing || !drawCanvasRef.current) return;

		const canvas = drawCanvasRef.current;
		const context = canvas.getContext("2d");
		if (!context) return;

		context.closePath();
		setIsDrawing(false);

		// Save the drawing canvas state to your history for undo/redo functionality
		const currentState = context.getImageData(
			0,
			0,
			canvas.width,
			canvas.height
		);
		const newHistory = drawingHistory.slice(0, historyIndex + 1);
		setDrawingHistory([...newHistory, currentState]);
		setHistoryIndex(newHistory.length);
	};

	// Undo function: Restores the previous state
	const undo = () => {
		if (historyIndex > 0 && drawCanvasRef.current) {
			const newIndex = historyIndex - 1;
			setHistoryIndex(newIndex);
			const context = drawCanvasRef.current.getContext("2d");
			if (context) {
				context.putImageData(drawingHistory[newIndex], 0, 0);
			}
		}
	};

	// Redo function: Restores the next state if available
	const redo = () => {
		if (historyIndex < drawingHistory.length - 1 && drawCanvasRef.current) {
			const newIndex = historyIndex + 1;
			setHistoryIndex(newIndex);
			const context = drawCanvasRef.current.getContext("2d");
			if (context) {
				context.putImageData(drawingHistory[newIndex], 0, 0);
			}
		}
	};
	// Handle post submission
	const handleSubmit = () => {
		if (!selectedImage) return;

		setIsSubmitting(true);

		// Get the final image from canvas
		let finalImage = selectedImage;
		if (canvasRef.current) {
			finalImage = canvasRef.current.toDataURL("image/png");
		}

		// Create a new post object
		const newPost = {
			id: Math.floor(Math.random() * 10000),
			content: caption,
			imageUrl: finalImage,
			createdAt: new Date(),
			likes: 0,
			comments: 0,
			author: {
				name: "Current User",
				username: "currentuser",
				avatar: "/placeholder.svg?height=40&width=40",
			},
		};

		// Simulate API call delay
		setTimeout(() => {
			// Add the post to the wall
			onPostCreated(newPost);

			// Reset form and close modal
			setSelectedImage(null);
			setCaption("");
			setActiveTab("upload");
			setIsSubmitting(false);
			onClose();
		}, 1000);
	};

	// Reset state when modal closes
	const handleClose = () => {
		setSelectedImage(null);
		setCaption("");
		setActiveTab("upload");
		onClose();
	};

	const handleNext = () => {
		const bgCanvas = bgCanvasRef.current;
		const drawCanvas = drawCanvasRef.current;

		if (!bgCanvas || !drawCanvas) return;
		const offscreenCanvas = document.createElement("canvas");
		offscreenCanvas.width = bgCanvas.width;
		offscreenCanvas.height = bgCanvas.height;
		const offscreenCtx = offscreenCanvas.getContext("2d");

		if (offscreenCtx) {
			// Draw the background layer first...
			offscreenCtx.drawImage(bgCanvas, 0, 0);
			// ...then the drawing layer on top
			offscreenCtx.drawImage(drawCanvas, 0, 0);
			// Store the combined result as a data URL
			setCompositeDataUrl(offscreenCanvas.toDataURL("image/png"));
		}
	};

	// Functions for Media URL
	const getEmbedUrl = (url: string) => {
		// YouTube
		if (url.includes("youtube.com") || url.includes("youtu.be")) {
			const videoId = url.includes("youtube.com")
				? url.split("v=")[1]?.split("&")[0]
				: url.split("youtu.be/")[1]?.split("?")[0];
			return videoId ? `https://www.youtube.com/embed/${videoId}` : null;
		}
		// TikTok
		else if (url.includes("tiktok.com")) {
			// For TikTok, we'd normally use their embed API, but for demo purposes:
			return url;
		}
		// Other media types could be added here
		return url;
	};

	const handlePreview = () => {
		if (!mediaUrl) {
			toast("Event has been created", {
				description: "Sunday, December 03, 2023 at 9:00 AM",
				action: {
					label: "Undo",
					onClick: () => console.log("Undo"),
				},
			});
			return;
		}

		setIsPreviewLoaded(true);
	};

	const handleReset = () => {
		setMediaUrl("");
		setCaption("");
		setIsPreviewLoaded(false);
	};

	const isYouTube =
		mediaUrl.includes("youtube.com") || mediaUrl.includes("youtu.be");
	const embedUrl = getEmbedUrl(mediaUrl);

	return (
		<Dialog open={isOpen} onOpenChange={handleClose}>
			<DialogContent className="sm:max-w-[90%] md:max-w-[800px] max-h-[90vh] overflow-y-auto bg-background border-2 border-primary/20">
				<DialogHeader>
					<DialogTitle className="text-2xl font-graffiti">
						{activeTab === "upload"
							? "Add New Post"
							: activeTab === "draw"
							? "Edit Your Image"
							: "Add Details"}
					</DialogTitle>
				</DialogHeader>

				{/* Step navigation */}
				{selectedImage && (
					<div className="flex items-center mb-4">
						<Button
							variant="ghost"
							size="sm"
							className="flex items-center gap-1"
							onClick={() =>
								setActiveTab(activeTab === "draw" ? "upload" : "draw")
							}
						>
							<ChevronLeft className="h-4 w-4" />
							{activeTab === "details" ? "Back to Drawing" : "Back"}
						</Button>

						<div className="flex-1 flex justify-center">
							<div className="flex items-center gap-2">
								<div
									className={`h-2 w-2 rounded-full ${
										activeTab === "upload" ? "bg-primary" : "bg-muted"
									}`}
								></div>
								<div
									className={`h-2 w-2 rounded-full ${
										activeTab === "draw" ? "bg-primary" : "bg-muted"
									}`}
								></div>
								<div
									className={`h-2 w-2 rounded-full ${
										activeTab === "details" ? "bg-primary" : "bg-muted"
									}`}
								></div>
							</div>
						</div>

						{activeTab === "draw" && (
							<Button
								variant="ghost"
								size="sm"
								onClick={() => setActiveTab("details")}
							>
								Next
							</Button>
						)}
					</div>
				)}

				{/* Upload Tab */}
				{activeTab === "upload" && (
					<div className="grid gap-6">
						<Tabs defaultValue="upload" className="w-full">
							<TabsList className="grid w-full grid-cols-2">
								<TabsTrigger value="upload">Upload</TabsTrigger>
								<TabsTrigger value="camera">Embed Link</TabsTrigger>
							</TabsList>

							<TabsContent value="upload" className="mt-4">
								<div className="border-2 border-dashed border-primary/20 rounded-md p-8 text-center">
									<div className="flex flex-col items-center gap-3">
										<ImageIcon className="h-12 w-12 text-muted-foreground" />
										<div className="text-sm text-muted-foreground">
											Drag and drop an image or click to browse
										</div>
										<div className="flex gap-2">
											<Button
												variant="outline"
												size="sm"
												className="mt-2 cursor-pointer relative"
											>
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

							<TabsContent value="camera" className="mt-4">
								<div className="border-2 border-primary/20 rounded-md p-8 text-center bg-black/5">
									<div className="flex flex-col items-center gap-3">
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

										{isPreviewLoaded && embedUrl && (
											<div className="space-y-2">
												<div className="flex items-center justify-between">
													<Label>Preview</Label>
													<Button
														type="button"
														variant="ghost"
														size="icon"
														onClick={handleReset}
														className="h-6 w-6"
													>
														<X className="h-4 w-4" />
													</Button>
												</div>
												<div className="rounded-md overflow-hidden border bg-muted">
													{isYouTube ? (
														<div className="aspect-video">
															<iframe
																src={embedUrl}
																className="w-full h-full"
																allowFullScreen
																allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
															></iframe>
														</div>
													) : (
														<div className="p-4 text-center text-sm text-muted-foreground">
															<p>Preview for {mediaUrl}</p>
															<p className="mt-2">
																(Non-YouTube previews would be implemented based
																on specific platform APIs)
															</p>
														</div>
													)}
												</div>
											</div>
										)}

										{isPreviewLoaded && (
											<div className="space-y-2">
												<Label htmlFor="caption">Caption</Label>
												<Textarea
													id="caption"
													placeholder="Add a caption to your media..."
													value={caption}
													onChange={(e) => setCaption(e.target.value)}
													rows={3}
												/>
											</div>
										)}
									</div>
								</div>
							</TabsContent>
						</Tabs>
					</div>
				)}

				{/* Drawing Tab */}
				{activeTab === "draw" && selectedImage && (
					<div className="grid gap-4 md:grid-cols-[1fr_250px]">
						{/* Canvas Area */}
						<div
							className="relative border rounded-md overflow-hidden bg-white"
							style={{
								width: selectedImage
									? drawCanvasRef.current?.width || 600
									: "auto",
								height: selectedImage
									? drawCanvasRef.current?.height || 400
									: "auto",
							}}
						>
							<canvas
								ref={bgCanvasRef}
								className="absolute inset-0"
								style={{zIndex: 0}}
							/>
							<canvas
								ref={drawCanvasRef}
								className="absolute inset-0"
								style={{zIndex: 1}}
								onMouseDown={startDrawing}
								onMouseMove={draw}
								onMouseUp={stopDrawing}
								onMouseLeave={stopDrawing}
							/>
						</div>

						{/* Drawing Tools */}
						<div className="bg-black/5 backdrop-blur-sm rounded-xl p-4">
							<div className="grid gap-4">
								<div className="flex flex-wrap gap-2 justify-center">
									<Button
										variant={drawingMode === "brush" ? "default" : "outline"}
										size="icon"
										onClick={() => setDrawingMode("brush")}
									>
										<Brush className="h-5 w-5" />
									</Button>
									<Button
										variant={drawingMode === "eraser" ? "default" : "outline"}
										size="icon"
										onClick={() => setDrawingMode("eraser")}
									>
										<Eraser className="h-5 w-5" />
									</Button>
									<Button
										variant="outline"
										size="icon"
										onClick={undo}
										disabled={historyIndex <= 0}
									>
										<Undo className="h-5 w-5" />
									</Button>
									<Button
										variant="outline"
										size="icon"
										onClick={redo}
										disabled={historyIndex >= drawingHistory.length - 1}
									>
										<Redo className="h-5 w-5" />
									</Button>
								</div>

								<div className="space-y-2">
									<div className="flex justify-between items-center">
										<Label>Brush Size</Label>
										<span className="text-sm">{brushSize}px</span>
									</div>
									<Slider
										value={[brushSize]}
										min={1}
										max={50}
										step={1}
										onValueChange={(value: React.SetStateAction<number>[]) =>
											setBrushSize(value[0])
										}
									/>
								</div>

								<div>
									<Label className="mb-2 block">Color</Label>
									<div className="flex flex-wrap gap-2">
										{colors.map((color) => (
											<button
												key={color}
												className={`h-8 w-8 rounded-full ${
													brushColor === color
														? "ring-2 ring-offset-2 ring-primary"
														: "ring-2 ring-offset-1 ring-gray-700"
												}`}
												style={{backgroundColor: color}}
												onClick={() => setBrushColor(color)}
											/>
										))}
									</div>
								</div>

								<div className="grid grid-cols-2 gap-2 mt-4">
									<Button
										variant="outline"
										size="sm"
										onClick={() => setActiveTab("upload")}
									>
										<X className="h-4 w-4 mr-2" />
										Cancel
									</Button>
									<Button
										size="sm"
										onClick={() => {
											setActiveTab("details");
											handleNext();
										}}
									>
										<Check className="h-4 w-4 mr-2" />
										Next
									</Button>
								</div>
							</div>
						</div>
					</div>
				)}

				{/* Details Tab */}
				{activeTab === "details" && selectedImage && (
					<div className="grid gap-6 md:grid-cols-[1fr_1fr]">
						<div className="relative border rounded-md overflow-hidden bg-white">
							{compositeDataUrl ? (
								// eslint-disable-next-line @next/next/no-img-element
								<img
									src={compositeDataUrl}
									alt="Combined Preview"
									className="max-w-full h-auto"
								/>
							) : (
								<p>Loading preview...</p>
							)}
						</div>

						<div className="space-y-4">
							<div className="space-y-2">
								<Label htmlFor="caption">Caption</Label>
								<Textarea
									id="caption"
									placeholder="Write a caption for your post..."
									value={caption}
									onChange={(e) => setCaption(e.target.value)}
									className="resize-none h-32"
								/>
							</div>

							<div className="pt-4">
								<Button
									className="w-full"
									onClick={handleSubmit}
									disabled={isSubmitting}
								>
									{isSubmitting ? (
										<div className="flex items-center gap-2">
											<div className="h-4 w-4 rounded-full border-2 border-t-transparent border-white animate-spin"></div>
											Posting...
										</div>
									) : (
										<>
											<Save className="h-4 w-4 mr-2" />
											Post to Wall
										</>
									)}
								</Button>
							</div>
						</div>
					</div>
				)}

				<DialogFooter>
					{activeTab === "upload" && (
						<Button variant="outline" onClick={handleClose}>
							Cancel
						</Button>
					)}
				</DialogFooter>
			</DialogContent>
		</Dialog>
	);
}
