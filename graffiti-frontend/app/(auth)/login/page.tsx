"use client";

import {useState} from "react";
import {Button} from "@/components/ui/button";
import {Input} from "@/components/ui/input";
import {
	Card,
	CardContent,
	CardDescription,
	CardFooter,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import {Tabs, TabsContent, TabsList, TabsTrigger} from "@/components/ui/tabs";
import {toast} from "sonner";
import {useRouter} from "next/navigation";
import {Eye, EyeClosed} from "lucide-react";

export default function Login() {
	const router = useRouter();
	const [tabValue, setTabValue] = useState("login");
	const [isLoading, setIsLoading] = useState(false);
	const [registerData, setRegisterData] = useState({
		username: "",
		fullname: "",
		email: "",
		password: "",
		confirmPassword: "",
	});
	const [loginData, setLoginData] = useState({
		email: "",
		password: "",
	});

	const [loginPasswordSeen, setLoginPasswordSeen] = useState(true);
	const [registerPasswordSeen, setRegisterPasswordSeen] = useState(true);
	const [confirmRegisterPasswordSeen, setConfirmRegisterPasswordSeen] =
		useState(true);

	const handleLogin = async (e: React.FormEvent) => {
		e.preventDefault();
		setIsLoading(true);
		const {email, password} = loginData;
		if (email === "" || password === "") {
			alert("Please fill in all fields");
			setIsLoading(false);
			return;
		}

		try {
			const res = await fetch(
				`${process.env.NEXT_PUBLIC_API_URL}/api/v1/auth/login`,
				{
					method: "POST",
					headers: {
						"Content-Type": "application/json",
					},
					body: JSON.stringify({email, password}),
					credentials: "include",
				}
			);

			const data = await res.json();

			if (!res.ok) {
				// const error = await res.json
				throw new Error(data.error || "Login failed");
			}
			toast.success("Login successful!");
			router.push("/");
		} catch (err) {
			toast.warning("Something wrong happened", {
				description: (err as Error).message,
			});
		} finally {
			setIsLoading(false);
			setLoginData({
				email: "",
				password: "",
			});
		}
	};

	const handleRegister = async (e: React.FormEvent) => {
		e.preventDefault();
		const {username, fullname, email, password, confirmPassword} = registerData;
		if (
			username === "" ||
			fullname === "" ||
			email === "" ||
			password === "" ||
			confirmPassword === ""
		) {
			alert("Please fill in all fields");
			return;
		}

		if (!email.includes("@")) {
			alert("Invalid email");
			return;
		}

		if (password !== confirmPassword) {
			alert("Passwords do not match");
			return;
		}

		if (password.length < 6) {
			alert("Password must be at least 6 characters");
			return;
		}

		if (username.length < 3) {
			alert("Username must be at least 3 characters");
			return;
		}

		if (/\s/.test(username)) {
			alert("Username must not contain spaces");
			return;
		}

		try {
			setIsLoading(true);
			const res = await fetch(
				`${process.env.NEXT_PUBLIC_API_URL}/api/v1/auth/register`,
				{
					method: "POST",
					headers: {
						"Content-Type": "application/json",
					},
					body: JSON.stringify({username, fullname, email, password}),
				}
			);

			const data = await res.json();
			if (!res.ok) {
				throw new Error(data.message || "Registration failed");
			}
			toast.success("Registration successful!");
		} catch (err) {
			console.error(err);
			toast.warning("Something wrong happened", {
				description: (err as Error).message,
			});
		} finally {
			setIsLoading(false);
			setTabValue("login");
			setRegisterData({
				username: "",
				fullname: "",
				email: "",
				password: "",
				confirmPassword: "",
			});
		}
	};

	const handleKeyPress = (e: React.KeyboardEvent<HTMLFormElement>) => {
		if (e.key === "Enter") {
			if (tabValue === "login") {
				handleLogin(e as unknown as React.FormEvent);
			} else {
				handleRegister(e as unknown as React.FormEvent);
			}
		}
	};

	return (
		<main className="min-h-screen bg-gradient-to-br from-purple-500 via-pink-500 to-orange-500 flex items-center justify-center p-4">
			<div className="absolute inset-0 bg-[url('/mockbg.jpg')] bg-cover bg-center opacity-10"></div>

			<div className="w-full max-w-md z-10 ">
				<h1 className="text-4xl font-bold text-center text-white mb-2">
					Graffiti
				</h1>
				<p className="text-center text-white/90 mb-8 text-lg">
					Express yourself without boundaries
				</p>

				<Card className="backdrop-blur-xl bg-white/90">
					<CardHeader>
						<CardTitle className="text-black text-2xl pt-3">
							Welcome to Graffiti
						</CardTitle>
						<CardDescription className="text-sm text-gray-500">
							Sign in to start creating or join the community
						</CardDescription>
					</CardHeader>
					<CardContent>
						<Tabs
							defaultValue="login"
							value={tabValue}
							onValueChange={(value) => setTabValue(value)}
							className="w-full"
						>
							<TabsList className="grid w-full grid-cols-2 mb-4 bg-gray-100 rounded-md">
								<TabsTrigger value="login" className="cursor-pointer">
									Login
								</TabsTrigger>
								<TabsTrigger value="register" className="cursor-pointer">
									Register
								</TabsTrigger>
							</TabsList>

							<TabsContent value="login">
								<form
									onSubmit={handleLogin}
									onKeyDown={handleKeyPress}
									className="text-black"
								>
									<div className="space-y-4">
										<Input
											type="email"
											placeholder="Email"
											required
											value={loginData.email}
											onChange={(e) => {
												setLoginData({
													...loginData,
													email: e.target.value,
												});
											}}
										/>
										<div className="flex gap-2 items-center">
											<Input
												type={loginPasswordSeen ? "password" : "text"}
												placeholder="Password"
												required
												value={loginData.password}
												onChange={(e) => {
													setLoginData({
														...loginData,
														password: e.target.value,
													});
												}}
											/>
											<Button
												variant={"ghost"}
												onClick={(e) => {
													e.preventDefault();
													setLoginPasswordSeen(!loginPasswordSeen);
												}}
											>
												{loginPasswordSeen ? <Eye /> : <EyeClosed />}
											</Button>
										</div>

										<Button
											variant={"special"}
											className="w-full"
											type="submit"
											disabled={isLoading}
										>
											{isLoading ? "Signing in..." : "Sign in"}
										</Button>
									</div>
								</form>
							</TabsContent>

							<TabsContent value="register">
								<form
									onSubmit={handleRegister}
									onKeyDown={handleKeyPress}
									className="text-black"
								>
									<div className="space-y-4">
										<Input
											type="email"
											placeholder="Email"
											required
											value={registerData.email}
											onChange={(e) => {
												setRegisterData({
													...registerData,
													email: e.target.value,
												});
											}}
										/>
										<Input
											type="text"
											placeholder="Username"
											required
											value={registerData.username}
											onChange={(e) =>
												setRegisterData({
													...registerData,
													username: e.target.value,
												})
											}
										/>
										<Input
											type="text"
											placeholder="Full Name"
											required
											value={registerData.fullname}
											onChange={(e) =>
												setRegisterData({
													...registerData,
													fullname: e.target.value,
												})
											}
										/>
										<div className="flex gap-2 items-center">
											<Input
												type={registerPasswordSeen ? "password" : "text"}
												placeholder="Password"
												required
												value={registerData.password}
												onChange={(e) => {
													setRegisterData({
														...registerData,
														password: e.target.value,
													});
												}}
											/>
											<Button
												variant={"ghost"}
												onClick={(e) => {
													e.preventDefault();
													setRegisterPasswordSeen(!registerPasswordSeen);
												}}
											>
												{registerPasswordSeen ? <Eye /> : <EyeClosed />}
											</Button>
										</div>
										<div className="flex gap-2 items-center">
											<Input
												type={confirmRegisterPasswordSeen ? "password" : "text"}
												placeholder="Confirm Password"
												required
												value={registerData.confirmPassword}
												onChange={(e) => {
													setRegisterData({
														...registerData,
														confirmPassword: e.target.value,
													});
												}}
											/>
											<Button
												variant={"ghost"}
												onClick={(e) => {
													e.preventDefault();
													setConfirmRegisterPasswordSeen(
														!confirmRegisterPasswordSeen
													);
												}}
											>
												{confirmRegisterPasswordSeen ? <Eye /> : <EyeClosed />}
											</Button>
										</div>
										<Button
											className="w-full "
											variant={"special"}
											type="submit"
											disabled={isLoading}
										>
											{isLoading ? "Creating account..." : "Create account"}
										</Button>
									</div>
								</form>
							</TabsContent>
						</Tabs>

						<div className="relative my-6">
							<div className="absolute inset-0 flex items-center">
								<div className="w-full border-t border-gray-300"></div>
							</div>
							<div className="relative flex justify-center text-sm">
								<span className="px-2 bg-white/50 text-gray-500">
									Or continue with
								</span>
							</div>
						</div>

						<div className="grid grid-cols-1 gap-3">
							<Button variant="outline" disabled className="w-full">
								More sign-in options coming soon
							</Button>
						</div>
					</CardContent>
					<CardFooter className="flex flex-col space-y-4 pb-4">
						<div className="text-sm text-center text-gray-500">
							By continuing, you agree to our Terms of Service and Privacy
							Policy
						</div>
					</CardFooter>
				</Card>
			</div>
		</main>
	);
}
