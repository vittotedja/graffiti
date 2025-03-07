"use client";

import type React from "react";

import {useState, useRef, useEffect} from "react";
import {
	Upload,
	ImageIcon,
	Camera,
	Brush,
	Eraser,
	Save,
	Undo,
	Redo,
	Check,
	X,
	ChevronLeft,
} from "lucide-react";
import NextImage from "next/image";

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

	const canvasRef = useRef<HTMLCanvasElement>(null);
	const contextRef = useRef<CanvasRenderingContext2D | null>(null);

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

	// Initialize canvas when image is selected
	useEffect(() => {
		if (selectedImage && canvasRef.current) {
			const canvas = canvasRef.current;
			const context = canvas.getContext("2d");

			if (context) {
				contextRef.current = context;

				// Load the image onto canvas
				const img = new Image();
				img.crossOrigin = "anonymous";
				img.onload = () => {
					// Set canvas dimensions to match image
					canvas.width = img.width;
					canvas.height = img.height;

					// Draw image on canvas
					context.drawImage(img, 0, 0, canvas.width, canvas.height);

					// Save initial state to history
					const initialState = context.getImageData(
						0,
						0,
						canvas.width,
						canvas.height
					);
					setDrawingHistory([initialState]);
					setHistoryIndex(0);
				};
				img.src = selectedImage;
			}
		}
	}, [selectedImage]);

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

	// Drawing functions
	const startDrawing = (e: React.MouseEvent<HTMLCanvasElement>) => {
		if (!contextRef.current) return;

		const canvas = canvasRef.current;
		if (!canvas) return;

		const rect = canvas.getBoundingClientRect();
		const x = (e.clientX - rect.left) * (canvas.width / rect.width);
		const y = (e.clientY - rect.top) * (canvas.height / rect.height);

		contextRef.current.beginPath();
		contextRef.current.moveTo(x, y);

		if (drawingMode === "brush") {
			contextRef.current.strokeStyle = brushColor;
			contextRef.current.lineWidth = brushSize;
			contextRef.current.lineCap = "round";
		} else if (drawingMode === "eraser") {
			contextRef.current.strokeStyle = "#ffffff";
			contextRef.current.lineWidth = brushSize;
			contextRef.current.lineCap = "round";
		}

		setIsDrawing(true);
	};

	const draw = (e: React.MouseEvent<HTMLCanvasElement>) => {
		if (!isDrawing || !contextRef.current || !canvasRef.current) return;

		const canvas = canvasRef.current;
		const rect = canvas.getBoundingClientRect();
		const x = (e.clientX - rect.left) * (canvas.width / rect.width);
		const y = (e.clientY - rect.top) * (canvas.height / rect.height);

		contextRef.current.lineTo(x, y);
		contextRef.current.stroke();
	};

	const stopDrawing = () => {
		if (!isDrawing || !contextRef.current || !canvasRef.current) return;

		contextRef.current.closePath();
		setIsDrawing(false);

		// Save current state to history
		const canvas = canvasRef.current;
		const currentState = contextRef.current.getImageData(
			0,
			0,
			canvas.width,
			canvas.height
		);

		// Remove any states after current index (if we've undone and then drawn something new)
		const newHistory = drawingHistory.slice(0, historyIndex + 1);
		setDrawingHistory([...newHistory, currentState]);
		setHistoryIndex(newHistory.length);
	};

	const undo = () => {
		if (historyIndex > 0 && contextRef.current && canvasRef.current) {
			setHistoryIndex(historyIndex - 1);
			contextRef.current.putImageData(drawingHistory[historyIndex - 1], 0, 0);
		}
	};

	const redo = () => {
		if (
			historyIndex < drawingHistory.length - 1 &&
			contextRef.current &&
			canvasRef.current
		) {
			setHistoryIndex(historyIndex + 1);
			contextRef.current.putImageData(drawingHistory[historyIndex + 1], 0, 0);
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
								<TabsTrigger value="camera">Camera</TabsTrigger>
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
												// as="label"
												// htmlFor="file-upload"
												variant="outline"
												size="sm"
												className="mt-2 cursor-pointer"
											>
												<Upload className="h-4 w-4 mr-2" />
												Upload Image
												<input
													id="file-upload"
													type="file"
													accept="image/*"
													className="hidden"
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
										<Camera className="h-12 w-12 text-muted-foreground" />
										<div className="text-sm text-muted-foreground">
											Take a photo with your camera
										</div>
										<Button
											variant="outline"
											size="sm"
											className="mt-2"
											onClick={handleMockUpload}
										>
											<Camera className="h-4 w-4 mr-2" />
											Simulate Camera
										</Button>
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
						<div className="relative border rounded-md overflow-hidden bg-white">
							<div
								className="relative w-full"
								style={{maxHeight: "60vh", overflow: "auto"}}
							>
								<canvas
									ref={canvasRef}
									onMouseDown={startDrawing}
									onMouseMove={draw}
									onMouseUp={stopDrawing}
									onMouseLeave={stopDrawing}
									className="max-w-full h-auto"
								/>
							</div>
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
														: ""
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
									<Button size="sm" onClick={() => setActiveTab("details")}>
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
							{canvasRef.current ? (
								<div
									className="relative w-full"
									style={{maxHeight: "60vh", overflow: "auto"}}
								>
									<canvas ref={canvasRef} className="max-w-full h-auto" />
								</div>
							) : (
								<NextImage
									src={selectedImage || "/placeholder.svg"}
									alt="Preview"
									width={600}
									height={600}
									className="max-w-full h-auto"
								/>
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
