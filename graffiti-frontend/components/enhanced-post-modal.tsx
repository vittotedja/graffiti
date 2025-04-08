"use client";

import React, {useState, useRef, useEffect, useCallback} from "react";
import {
	Dialog,
	DialogContent,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";
import {Button} from "@/components/ui/button";
import {ChevronLeft} from "lucide-react";
import {toast} from "sonner";
import {fetchWithAuth} from "@/lib/auth";
import {useUser} from "@/hooks/useUser";

import UploadTab from "./upload-tab";
import DrawTab from "./draw-tab";
import DetailsTab from "./details-tab";
import {getPresignedUrl, uploadToS3} from "@/lib/s3-uploader";
import {RequestPost, Platform} from "@/types/post";

interface EnhancedPostModalProps {
	isOpen: boolean;
	onClose: () => void;
	wallId: string;
	onPostCreated: () => void;
}

export function EnhancedPostModal({
	isOpen,
	onClose,
	wallId,
	onPostCreated,
}: EnhancedPostModalProps) {
	const {user} = useUser();
	const [activeTab, setActiveTab] = useState("upload");
	const [selectedImage, setSelectedImage] = useState<string | null>(null);
	const [originalImage, setOriginalImage] = useState<HTMLImageElement | null>(
		null
	);
	const [caption, setCaption] = useState("");
	const [mediaUrl, setMediaUrl] = useState("");
	const [tiktokHtml, setTiktokHtml] = useState<string | null>(null);
	const [isPreviewLoaded, setIsPreviewLoaded] = useState(false);
	const [compositeDataUrl, setCompositeDataUrl] = useState<string | null>(null);
	const [isSubmitting, setIsSubmitting] = useState(false);

	// Canvas references
	const bgCanvasRef = useRef<HTMLCanvasElement>(null);
	const drawCanvasRef = useRef<HTMLCanvasElement>(null);

	// Drawing state
	const [isDrawing, setIsDrawing] = useState(false);
	const [drawingMode, setDrawingMode] = useState("brush");
	const [brushSize, setBrushSize] = useState(5);
	const [brushColor, setBrushColor] = useState("#ff3b30");
	const [drawingHistory, setDrawingHistory] = useState<ImageData[]>([]);
	const [historyIndex, setHistoryIndex] = useState(-1);

	const colors = [
		"#ff3b30",
		"#ff9500",
		"#ffcc00",
		"#34c759",
		"#5ac8fa",
		"#007aff",
		"#5856d6",
		"#af52de",
		"#000000",
		"#ffffff",
	];

	// Generate composite image for preview - moderate quality for UI display
	const generateComposite = useCallback(() => {
		if (!bgCanvasRef.current || !drawCanvasRef.current) return;
		const offscreenCanvas = document.createElement("canvas");
		offscreenCanvas.width = bgCanvasRef.current.width;
		offscreenCanvas.height = bgCanvasRef.current.height;
		const offscreenCtx = offscreenCanvas.getContext("2d");
		if (offscreenCtx) {
			// Enable image smoothing for better preview quality
			offscreenCtx.imageSmoothingEnabled = true;

			// Draw the background and drawing layers
			offscreenCtx.drawImage(bgCanvasRef.current, 0, 0);
			offscreenCtx.drawImage(drawCanvasRef.current, 0, 0);

			// Generate data URL at moderate quality (better for preview performance)
			const dataUrl = offscreenCanvas.toDataURL("image/jpeg", 0.85);
			setCompositeDataUrl(dataUrl);
		}
	}, []);

	useEffect(() => {
		if (selectedImage && bgCanvasRef.current && drawCanvasRef.current) {
			const bgCanvas = bgCanvasRef.current;
			const drawCanvas = drawCanvasRef.current;
			const bgContext = bgCanvas.getContext("2d");
			const drawContext = drawCanvas.getContext("2d");
			if (!user || !bgContext || !drawContext) return;

			const MAX_WIDTH = 480;
			const MAX_HEIGHT = 480;

			if (selectedImage.includes("placeholder.svg")) {
				bgCanvas.width = MAX_WIDTH;
				bgCanvas.height = MAX_HEIGHT;
				drawCanvas.width = MAX_WIDTH;
				drawCanvas.height = MAX_HEIGHT;

				bgContext.fillStyle = "#ffffff";
				bgContext.fillRect(0, 0, MAX_WIDTH, MAX_HEIGHT);

				drawContext.clearRect(0, 0, MAX_WIDTH, MAX_HEIGHT);
				const initialState = drawContext.getImageData(
					0,
					0,
					MAX_WIDTH,
					MAX_HEIGHT
				);
				setDrawingHistory([initialState]);
				setHistoryIndex(0);
			} else {
				const img = new Image();
				img.crossOrigin = "anonymous";
				img.onload = () => {
					// Store the original image for high-quality export later
					setOriginalImage(img);

					// Calculate dimensions for the editing canvas (keep reasonable size for editing)
					let width = img.width;
					let height = img.height;

					// Use the original MAX_WIDTH/MAX_HEIGHT for editing UX
					const scale = Math.min(MAX_WIDTH / width, MAX_HEIGHT / height, 1);
					width = width * scale;
					height = height * scale;

					// Set canvas dimensions to a reasonable size for editing
					bgCanvas.width = width;
					bgCanvas.height = height;
					drawCanvas.width = width;
					drawCanvas.height = height;

					// Enable image rendering optimized for editing
					bgContext.imageSmoothingEnabled = true;

					// Draw the image to the background canvas
					bgContext.drawImage(img, 0, 0, width, height);
					drawContext.clearRect(0, 0, width, height);
					const initialState = drawContext.getImageData(0, 0, width, height);
					setDrawingHistory([initialState]);
					setHistoryIndex(0);
				};
				img.src = selectedImage;
			}
		}
	}, [selectedImage, user]);

	// File upload handlers
	const handleFileUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
		const file = e.target.files?.[0];
		if (file) {
			// Store as a data URL for canvas display
			const reader = new FileReader();
			reader.onload = (event) => {
				setSelectedImage(event.target?.result as string);
				setActiveTab("draw");
			};
			reader.readAsDataURL(file);

			// Note: Original image is stored when image loads in the useEffect
		}
	};

	const handleMockUpload = () => {
		setSelectedImage("/placeholder.svg?height=600&width=600");
		setActiveTab("draw");
	};

	// Drawing handlers
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

	// Media preview functions
	const getEmbedUrl = (
		url: string
	): {embedUrl: string | null; type: Platform} => {
		if (url.includes("youtube.com") || url.includes("youtu.be")) {
			const videoId = url.includes("youtube.com")
				? url.split("v=")[1]?.split("&")[0]
				: url.split("youtu.be/")[1]?.split("?")[0];
			return {
				embedUrl: videoId ? `https://www.youtube.com/embed/${videoId}` : null,
				type: "youtube",
			};
		} else if (url.includes("spotify.com")) {
			if (
				url.includes("/track/") ||
				url.includes("/album/") ||
				url.includes("/playlist/")
			) {
				const parts = url.split(".com/");
				if (parts.length > 1) {
					const path = parts[1];
					return {
						embedUrl: `https://open.spotify.com/embed/${path}`,
						type: "spotify",
					};
				}
			}
			return {embedUrl: url, type: "spotify"};
		} else if (url.includes("tiktok.com")) {
			return {embedUrl: url, type: "tiktok"};
		}
		return {embedUrl: url, type: "others"};
	};

	const handlePreview = async () => {
		if (!mediaUrl) {
			toast.warning("No Media URL found", {
				description: "Please enter a valid URL",
			});
			return;
		}
		const {type} = getEmbedUrl(mediaUrl);
		if (type === "tiktok") {
			try {
				const response = await fetch(
					`https://www.tiktok.com/oembed?url=${mediaUrl}`
				);
				const data = await response.json();
				if (data.html) {
					setTiktokHtml(data.html);
					setIsPreviewLoaded(true);
					setTimeout(() => {
						if (!document.getElementById("tiktok-embed-script")) {
							const script = document.createElement("script");
							script.src = "https://www.tiktok.com/embed.js";
							script.async = true;
							script.id = "tiktok-embed-script";
							document.body.appendChild(script);
						}
					}, 500);
				} else {
					toast.error("TikTok embed failed", {
						description: "Could not fetch the TikTok embed.",
					});
				}
			} catch (error) {
				toast.error("TikTok fetch error", {
					description: "Failed to fetch the TikTok embed URL.",
				});
				console.error("Error fetching TikTok embed:", error);
			}
		} else {
			setIsPreviewLoaded(true);
		}
	};

	const handleReset = () => {
		setMediaUrl("");
		setCaption("");
		setIsPreviewLoaded(false);
	};

	const renderIframe = () => {
		const {embedUrl, type} = getEmbedUrl(mediaUrl);
		if (!embedUrl) return null;
		switch (type) {
			case "youtube":
				return (
					<iframe
						src={embedUrl}
						className="w-full h-full aspect-video"
						allowFullScreen
						allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
					></iframe>
				);
			case "spotify":
				return (
					<iframe
						className="border-r-[12px]"
						src={embedUrl}
						width="100%"
						height="352"
						allowFullScreen
						allow="autoplay; clipboard-write; encrypted-media; fullscreen; picture-in-picture"
						loading="lazy"
					></iframe>
				);
			case "tiktok":
				if (!tiktokHtml) return <p>Loading TikTok preview...</p>;
				return (
					<div className="w-full">
						<div dangerouslySetInnerHTML={{__html: tiktokHtml}} />
					</div>
				);
			default:
				return (
					<div className="p-4 text-center text-sm text-muted-foreground">
						<p>Preview for {mediaUrl}</p>
						<p className="mt-2">
							(Non-YouTube/Spotify previews would be implemented based on
							specific platform APIs)
						</p>
					</div>
				);
		}
	};

	// S3 Upload function with balanced quality/UX approach
	const handleUploadToS3 = async () => {
		if (!bgCanvasRef.current || !drawCanvasRef.current) return;

		try {
			// Create a high-quality canvas for the final upload
			const highQualityCanvas = document.createElement("canvas");
			const highQualityCtx = highQualityCanvas.getContext("2d");

			if (highQualityCtx) {
				// Determine the best dimensions for the final image
				let finalWidth, finalHeight;

				if (originalImage) {
					// If we have the original image, use its dimensions with reasonable limits
					const MAX_HQ_DIMENSION = 1200; // High quality but not excessive
					const originalWidth = originalImage.naturalWidth;
					const originalHeight = originalImage.naturalHeight;

					// Scale down if needed, but maintain quality
					const qualityScale = Math.min(
						MAX_HQ_DIMENSION / originalWidth,
						MAX_HQ_DIMENSION / originalHeight,
						1
					);

					finalWidth = Math.round(originalWidth * qualityScale);
					finalHeight = Math.round(originalHeight * qualityScale);

					// Set canvas size to the high-quality dimensions
					highQualityCanvas.width = finalWidth;
					highQualityCanvas.height = finalHeight;

					// Set high-quality rendering options
					highQualityCtx.imageSmoothingEnabled = true;
					highQualityCtx.imageSmoothingQuality = "high";

					// Draw the original image at high quality
					highQualityCtx.drawImage(
						originalImage,
						0,
						0,
						finalWidth,
						finalHeight
					);

					// Scale drawing canvas to match the high-quality dimensions
					const drawingScale = finalWidth / bgCanvasRef.current.width;

					// Apply the drawing layer scaled to match the high-quality canvas
					highQualityCtx.save();
					highQualityCtx.scale(drawingScale, drawingScale);
					highQualityCtx.drawImage(drawCanvasRef.current, 0, 0);
					highQualityCtx.restore();
				} else {
					// Fallback to the editing canvas if original isn't available
					// Still try to improve quality from the editing canvas
					highQualityCanvas.width = bgCanvasRef.current.width;
					highQualityCanvas.height = bgCanvasRef.current.height;

					highQualityCtx.imageSmoothingEnabled = true;
					highQualityCtx.imageSmoothingQuality = "high";

					highQualityCtx.drawImage(bgCanvasRef.current, 0, 0);
					highQualityCtx.drawImage(drawCanvasRef.current, 0, 0);
				}

				// Create a high-quality blob from the canvas
				const blob = await new Promise<Blob>((resolve, reject) => {
					highQualityCanvas.toBlob(
						(blob) => {
							if (blob) resolve(blob);
							else reject(new Error("Failed to generate image from canvas."));
						},
						"image/jpeg",
						0.92 // Good quality without excessive file size
					);
				});

				const filename = `image-${Date.now()}.jpg`;
				const presignedUrlData = await getPresignedUrl(filename, blob, "posts");
				await uploadToS3(presignedUrlData.presignedUrl, blob);
				toast.success("Image uploaded successfully!");
				return presignedUrlData.publicUrl;
			}
		} catch (error) {
			toast.error("Image upload failed.");
			console.error("S3 Upload Error:", error);
		}
	};

	const uploadPosts = async (postType: "media" | "embed") => {
		setIsSubmitting(true);

		let newPost: RequestPost;

		if (postType === "media") {
			const public_url = await handleUploadToS3();
			newPost = {
				post_type: "media",
				media_url: public_url || "",
				wall_id: wallId,
			};
		} else {
			const {embedUrl} = getEmbedUrl(mediaUrl);
			newPost = {
				post_type: "embed_link",
				media_url: embedUrl,
				wall_id: wallId,
			};
		}
		// TODO: Author to be set to the logged-in user
		try {
			if (newPost.media_url == "") return;
			const response = await fetchWithAuth("/api/v1/posts", {
				method: "POST",
				body: JSON.stringify({
					...newPost,
					author: user?.id,
				}),
			});

			if (!response.ok) throw new Error("Something went wrong");

			toast.success("Post created successfully");
		} catch (error) {
			console.error(error);
			toast.error("Something went wrong");
		} finally {
			setTimeout(() => {
				onPostCreated();
				setSelectedImage(null);
				setOriginalImage(null);
				setCaption("");
				setActiveTab("upload");
				setMediaUrl("");
				setIsPreviewLoaded(false);
				setCompositeDataUrl(null);
				setIsSubmitting(false);
				onClose();
			}, 1000);
		}
	};

	const handleSubmit = () => {
		if (mediaUrl && isPreviewLoaded) {
			uploadPosts("embed");
		} else if (selectedImage) {
			uploadPosts("media");
		}
	};

	const handleNext = () => {
		generateComposite();
		setActiveTab("details");
	};

	const handleClose = () => {
		setSelectedImage(null);
		setOriginalImage(null);
		setCaption("");
		setActiveTab("upload");
		onClose();
	};

	if (!user) return null;

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

				{/* Step Navigation */}
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
					</div>
				)}

				{/* Hide/Show Tab Components using CSS */}
				<div className={activeTab !== "upload" ? "hidden" : ""}>
					<UploadTab
						mediaUrl={mediaUrl}
						setMediaUrl={setMediaUrl}
						caption={caption}
						setCaption={setCaption}
						handleFileUpload={handleFileUpload}
						handleMockUpload={handleMockUpload}
						handlePreview={handlePreview}
						isPreviewLoaded={isPreviewLoaded}
						renderIframe={renderIframe}
						handleReset={handleReset}
						handleSubmit={handleSubmit}
						isSubmitting={isSubmitting}
					/>
				</div>
				<div className={activeTab !== "draw" ? "hidden" : ""}>
					{selectedImage && (
						<DrawTab
							selectedImage={selectedImage}
							bgCanvasRef={bgCanvasRef}
							drawCanvasRef={drawCanvasRef}
							startDrawing={startDrawing}
							draw={draw}
							stopDrawing={stopDrawing}
							drawingMode={drawingMode}
							setDrawingMode={setDrawingMode}
							brushSize={brushSize}
							setBrushSize={setBrushSize}
							brushColor={brushColor}
							setBrushColor={setBrushColor}
							colors={colors}
							undo={undo}
							redo={redo}
							setActiveTab={setActiveTab}
							handleNext={handleNext}
						/>
					)}
				</div>
				<div className={activeTab !== "details" ? "hidden" : ""}>
					{selectedImage && (
						<DetailsTab
							compositeDataUrl={compositeDataUrl}
							caption={caption}
							setCaption={setCaption}
							handleSubmit={handleSubmit}
							isSubmitting={isSubmitting}
						/>
					)}
				</div>

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
