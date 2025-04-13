"use client";

import React from "react";
import {Brush, Eraser, Undo, Redo, Check, X} from "lucide-react";
import {Button} from "@/components/ui/button";
import {Slider} from "@/components/ui/slider";
import {Label} from "@/components/ui/label";

interface DrawTabProps {
	selectedImage: string;
	bgCanvasRef: React.RefObject<HTMLCanvasElement | null>;
	drawCanvasRef: React.RefObject<HTMLCanvasElement | null>;
	startDrawing: (
		e: React.MouseEvent<HTMLCanvasElement> | React.TouchEvent<HTMLCanvasElement>
	) => void;
	draw: (
		e: React.MouseEvent<HTMLCanvasElement> | React.TouchEvent<HTMLCanvasElement>
	) => void;
	stopDrawing: () => void;
	drawingMode: string;
	setDrawingMode: (mode: string) => void;
	brushSize: number;
	setBrushSize: (size: number) => void;
	brushColor: string;
	setBrushColor: (color: string) => void;
	colors: string[];
	undo: () => void;
	redo: () => void;
	setActiveTab: (tab: string) => void;
	handleNext: () => void;
}

const DrawTab: React.FC<DrawTabProps> = ({
	// selectedImage,
	bgCanvasRef,
	drawCanvasRef,
	startDrawing,
	draw,
	stopDrawing,
	drawingMode,
	setDrawingMode,
	brushSize,
	setBrushSize,
	brushColor,
	setBrushColor,
	colors,
	undo,
	redo,
	setActiveTab,
	handleNext,
}) => {
	// Preset brush sizes for touch-friendly selection
	const brushPresets = [2, 5, 10, 20, 30];

	return (
		<div className="grid gap-4 md:grid-cols-[1fr_250px] sm:grid-cols-1">
			{/* Canvas Area */}
			<div
				className="relative border rounded-md overflow-hidden bg-white mx-auto"
				style={{
					width: "100%",
					maxWidth: drawCanvasRef.current?.width || 600,
					height: drawCanvasRef.current?.height || 400,
				}}
			>
				<canvas
					ref={bgCanvasRef}
					className="absolute inset-0"
					style={{zIndex: 0}}
				/>
				<canvas
					ref={drawCanvasRef}
					className="absolute inset-0 touch-none"
					style={{zIndex: 1}}
					onMouseDown={startDrawing}
					onMouseMove={draw}
					onMouseUp={stopDrawing}
					onMouseLeave={stopDrawing}
					onTouchStart={startDrawing}
					onTouchMove={draw}
					onTouchEnd={stopDrawing}
					onTouchCancel={stopDrawing}
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
						<Button variant="outline" size="icon" onClick={undo}>
							<Undo className="h-5 w-5" />
						</Button>
						<Button variant="outline" size="icon" onClick={redo}>
							<Redo className="h-5 w-5" />
						</Button>
					</div>

					{/* Touch-friendly brush size presets */}
					<div>
						<Label className="mb-2 block">Brush Size: {brushSize}px</Label>
						<div className="flex flex-wrap gap-2 justify-between mb-2">
							{brushPresets.map((size) => (
								<Button
									key={size}
									variant={brushSize === size ? "default" : "outline"}
									size="sm"
									className="flex-1"
									onClick={() => setBrushSize(size)}
								>
									{size}px
								</Button>
							))}
						</div>
						<Slider
							value={[brushSize]}
							min={1}
							max={50}
							step={1}
							onValueChange={(value: number[]) => setBrushSize(value[0])}
						/>
					</div>

					<div>
						<Label className="mb-2 block">Color</Label>
						<div className="flex flex-wrap gap-2">
							{colors.map((color) => (
								<button
									key={color}
									className={`h-10 w-10 rounded-full ${
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
						<Button size="sm" onClick={handleNext}>
							<Check className="h-4 w-4 mr-2" />
							Next
						</Button>
					</div>
				</div>
			</div>
		</div>
	);
};

export default DrawTab;
