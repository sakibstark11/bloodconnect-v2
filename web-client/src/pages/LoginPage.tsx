import { LoginForm } from "@/components/LoginForm"
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "@/components/ui/card"
import { Droplets } from "lucide-react"

export default function LoginPage() {
  return (
    <div className="min-h-screen flex items-center justify-center p-4 bg-[radial-gradient(ellipse_at_top,_var(--tw-gradient-stops))] from-destructive/20 via-background to-background">
      <div className="w-full max-w-md animate-in fade-in zoom-in duration-500">
        <div className="flex flex-col items-center gap-2 mb-8">
          <div className="h-16 w-16 rounded-3xl bg-destructive/10 flex items-center justify-center ring-1 ring-destructive/20 shadow-[0_0_20px_rgba(239,68,68,0.2)]">
            <Droplets className="h-10 w-10 text-destructive" />
          </div>
          <h1 className="text-4xl font-black tracking-tight mt-2">BloodConnect</h1>
          <p className="text-muted-foreground font-medium">Saving lives, one drop at a time.</p>
        </div>
        
        <Card className="border-none shadow-2xl bg-card/60 backdrop-blur-xl ring-1 ring-white/10">
          <CardHeader>
            <CardTitle>Welcome back</CardTitle>
            <CardDescription>Enter your credentials to access your account</CardDescription>
          </CardHeader>
          <CardContent>
            <LoginForm />
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
