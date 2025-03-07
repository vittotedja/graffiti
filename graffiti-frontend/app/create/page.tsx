"use client";

import {useState} from "react";
import Link from "next/link";
import {
	ChevronLeft,
	Upload,
	Brush,
	Eraser,
	Undo,
	Redo,
	ImageIcon,
	Type,
	Sticker,
	Save,
} from "lucide-react";

import {Button} from "@/components/ui/button";
import {Tabs, TabsContent, TabsList, TabsTrigger} from "@/components/ui/tabs";
import {Slider} from "@/components/ui/slider";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";
import {Input} from "@/components/ui/input";
import {Label} from "@/components/ui/label";

export default function CreatePostPage() {
	const [selectedTool, setSelectedTool] = useState("brush");
	const [brushSize, setBrushSize] = useState(5);
	const [brushColor, setBrushColor] = useState("#ff3b30");

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

	return (
		<div className="min-h-screen bg-[url('/images/concrete-texture.jpg')] bg-cover">
			<div className="container mx-auto px-4 pb-6">
				{/* Header */}
				<header className="py-4 flex items-center justify-between">
					<div className="flex items-center gap-2">
						<Link href="/">
							<Button variant="ghost" size="icon" className="rounded-full">
								<ChevronLeft className="h-6 w-6" />
							</Button>
						</Link>
						<h1 className="text-2xl font-bold font-graffiti">
							Create New Post
						</h1>
					</div>
					<Button className="bg-primary hover:bg-primary/90 text-white">
						Post
					</Button>
				</header>

				{/* Canvas Area */}
				<div className="bg-white rounded-xl overflow-hidden shadow-lg mb-6">
					<div className="relative aspect-square w-full">
						<div className="absolute inset-0 flex items-center justify-center bg-gray-100">
							<div className="text-center p-6">
								<div className="mb-4">
									<Upload className="h-12 w-12 mx-auto text-muted-foreground" />
								</div>
								<h3 className="text-lg font-medium mb-2">
									Upload an image to start
								</h3>
								<p className="text-muted-foreground text-sm mb-4">
									Drag and drop or click to upload
								</p>
								<Button>
									<ImageIcon className="h-4 w-4 mr-2" />
									Select Image
								</Button>
							</div>
						</div>
					</div>
				</div>

				{/* Drawing Tools */}
				<div className="bg-black/10 backdrop-blur-sm rounded-xl p-4 mb-6">
					<Tabs defaultValue="draw" className="w-full">
						<TabsList className="grid w-full grid-cols-3 mb-4">
							<TabsTrigger value="draw">Draw</TabsTrigger>
							<TabsTrigger value="text">Text</TabsTrigger>
							<TabsTrigger value="stickers">Stickers</TabsTrigger>
						</TabsList>

						<TabsContent value="draw" className="mt-0">
							<div className="grid gap-4">
								<div className="flex flex-wrap gap-2 justify-center">
									<Button
										variant={selectedTool === "brush" ? "default" : "outline"}
										size="icon"
										onClick={() => setSelectedTool("brush")}
									>
										<Brush className="h-5 w-5" />
									</Button>
									<Button
										variant={selectedTool === "eraser" ? "default" : "outline"}
										size="icon"
										onClick={() => setSelectedTool("eraser")}
									>
										<Eraser className="h-5 w-5" />
									</Button>
									<Button variant="outline" size="icon">
										<Undo className="h-5 w-5" />
									</Button>
									<Button variant="outline" size="icon">
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
										onValueChange={(value) => setBrushSize(value[0])}
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
							</div>
						</TabsContent>

						<TabsContent value="text" className="mt-0">
							<div className="grid gap-4">
								<div className="space-y-2">
									<Label htmlFor="text-input">Add Text</Label>
									<Input id="text-input" placeholder="Type something..." />
								</div>

								<div className="space-y-2">
									<Label htmlFor="font-select">Font Style</Label>
									<Select defaultValue="default">
										<SelectTrigger id="font-select">
											<SelectValue placeholder="Select font" />
										</SelectTrigger>
										<SelectContent>
											<SelectItem value="default">Default</SelectItem>
											<SelectItem value="graffiti">Graffiti</SelectItem>
											<SelectItem value="tag">Tag</SelectItem>
											<SelectItem value="bubble">Bubble</SelectItem>
											<SelectItem value="wildstyle">Wildstyle</SelectItem>
										</SelectContent>
									</Select>
								</div>

								<div className="space-y-2">
									<Label>Text Color</Label>
									<div className="flex flex-wrap gap-2">
										{colors.map((color) => (
											<button
												key={color}
												className={`h-8 w-8 rounded-full`}
												style={{backgroundColor: color}}
											/>
										))}
									</div>
								</div>

								<Button>
									<Type className="h-4 w-4 mr-2" />
									Add Text to Canvas
								</Button>
							</div>
						</TabsContent>

						<TabsContent value="stickers" className="mt-0">
							<div className="grid grid-cols-4 gap-2">
								{[1, 2, 3, 4, 5, 6, 7, 8].map((i) => (
									<div
										key={i}
										className="aspect-square bg-muted rounded-md flex items-center justify-center hover:bg-muted/80 cursor-pointer"
									>
										<Sticker className="h-8 w-8 text-muted-foreground" />
									</div>
								))}
							</div>
						</TabsContent>
					</Tabs>
				</div>

				{/* Post Details */}
				<div className="bg-black/10 backdrop-blur-sm rounded-xl p-4">
					<div className="space-y-4">
						<div className="space-y-2">
							<Label htmlFor="caption">Caption</Label>
							<Input id="caption" placeholder="Write a caption..." />
						</div>

						<div className="space-y-2">
							<Label htmlFor="wall-select">Add to Wall</Label>
							<Select defaultValue="birthday">
								<SelectTrigger id="wall-select">
									<SelectValue placeholder="Select wall" />
								</SelectTrigger>
								<SelectContent>
									<SelectItem value="birthday">Birthday Wall</SelectItem>
									<SelectItem value="graduation">Graduation</SelectItem>
									<SelectItem value="summer">Summer Trip</SelectItem>
									<SelectItem value="art">Art Projects</SelectItem>
								</SelectContent>
							</Select>
						</div>

						<Button className="w-full">
							<Save className="h-4 w-4 mr-2" />
							Save to Drafts
						</Button>
					</div>
				</div>
			</div>
		</div>
	);
}
