import { useState, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { User, Brain, Sparkles, RefreshCw } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import { useToast } from "@/components/ui/toast";
import { api } from "@/lib/api";

interface UserProfile {
  staticProfile: {
    traits: string[];
    preferences: string[];
    expertise: string[];
    summary: string;
  };
  dynamicProfile: {
    recentFocus: string[];
    goals: string[];
    activity: string;
    summary: string;
  };
  generatedAt: string;
}

export default function ProfilePage() {
  const { t } = useTranslation();
  const { toast } = useToast();
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchProfile = async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await api.axiosInstance.get<UserProfile>("/v1/profile");
      setProfile(data.data);
    } catch (err: any) {
      setError(err.message || t("profile.loadError"));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchProfile();
  }, []);

  return (
    <div className="min-h-screen bg-background">
      <header className="sticky top-0 z-10 border-b bg-gradient-to-r from-brain-primary to-brain-accent text-white">
        <div className="px-4 py-[14px]">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="p-1.5 bg-white/20 rounded-lg backdrop-blur-sm">
                <User className="h-5 w-5 text-white" />
              </div>
              <div>
                <h1 className="text-base font-bold leading-tight">{t("profile.title")}</h1>
                <p className="text-xs text-white/80">
                  {t("profile.subtitle")}
                </p>
              </div>
            </div>
            <Button variant="outline" size="sm" className="bg-white/20 text-white border-white/30 hover:bg-white/30" onClick={fetchProfile}>
              <RefreshCw className="h-4 w-4" />
            </Button>
          </div>
        </div>
      </header>

      <main className="container mx-auto px-4 py-8">
        {loading && (
          <div className="flex justify-center py-12">
            <Spinner size="lg" />
          </div>
        )}

        {error && (
          <Card>
            <CardContent className="p-8 text-center">
              <p className="text-destructive mb-4">{error}</p>
              <Button onClick={fetchProfile}>{t("profile.tryAgain")}</Button>
            </CardContent>
          </Card>
        )}

        {!loading && profile && (
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* Static Profile */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Brain className="h-5 w-5" />
                  {t("profile.static")}
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <p className="text-sm text-muted-foreground">{profile.staticProfile.summary}</p>

                {profile.staticProfile.traits.length > 0 && (
                  <div>
                    <h4 className="text-sm font-medium mb-2">{t("profile.traits")}</h4>
                    <div className="flex flex-wrap gap-2">
                      {profile.staticProfile.traits.map((trait, i) => (
                        <span key={i} className="px-2 py-1 bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300 text-xs rounded-full">
                          {trait}
                        </span>
                      ))}
                    </div>
                  </div>
                )}

                {profile.staticProfile.preferences.length > 0 && (
                  <div>
                    <h4 className="text-sm font-medium mb-2">{t("profile.preferences")}</h4>
                    <div className="flex flex-wrap gap-2">
                      {profile.staticProfile.preferences.map((pref, i) => (
                        <span key={i} className="px-2 py-1 bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-300 text-xs rounded-full">
                          {pref}
                        </span>
                      ))}
                    </div>
                  </div>
                )}

                {profile.staticProfile.expertise.length > 0 && (
                  <div>
                    <h4 className="text-sm font-medium mb-2">{t("profile.expertise")}</h4>
                    <div className="flex flex-wrap gap-2">
                      {profile.staticProfile.expertise.map((exp, i) => (
                        <span key={i} className="px-2 py-1 bg-purple-100 dark:bg-purple-900/30 text-purple-700 dark:text-purple-300 text-xs rounded-full">
                          {exp}
                        </span>
                      ))}
                    </div>
                  </div>
                )}
              </CardContent>
            </Card>

            {/* Dynamic Profile */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Sparkles className="h-5 w-5" />
                  {t("profile.dynamic")}
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <p className="text-sm text-muted-foreground">{profile.dynamicProfile.summary}</p>

                {profile.dynamicProfile.recentFocus.length > 0 && (
                  <div>
                    <h4 className="text-sm font-medium mb-2">{t("profile.recentFocus")}</h4>
                    <ul className="space-y-1">
                      {profile.dynamicProfile.recentFocus.map((focus, i) => (
                        <li key={i} className="text-sm text-muted-foreground flex items-start gap-2">
                          <span className="text-primary mt-1">•</span>
                          {focus}
                        </li>
                      ))}
                    </ul>
                  </div>
                )}

                {profile.dynamicProfile.goals.length > 0 && (
                  <div>
                    <h4 className="text-sm font-medium mb-2">{t("profile.goals")}</h4>
                    <ul className="space-y-1">
                      {profile.dynamicProfile.goals.map((goal, i) => (
                        <li key={i} className="text-sm text-muted-foreground flex items-start gap-2">
                          <span className="text-primary mt-1">•</span>
                          {goal}
                        </li>
                      ))}
                    </ul>
                  </div>
                )}

                {profile.dynamicProfile.activity && (
                  <div>
                    <h4 className="text-sm font-medium mb-2">{t("profile.activity")}</h4>
                    <p className="text-sm text-muted-foreground">{profile.dynamicProfile.activity}</p>
                  </div>
                )}
              </CardContent>
            </Card>
          </div>
        )}

        {!loading && !profile && !error && (
          <Card>
            <CardContent className="p-12 text-center">
              <User className="h-16 w-16 mx-auto mb-4 text-muted-foreground opacity-50" />
              <h3 className="text-lg font-semibold mb-2">{t("profile.empty")}</h3>
              <p className="text-muted-foreground mb-4">
                {t("profile.emptyDesc")}
              </p>
              <Button onClick={fetchProfile}>{t("profile.generate")}</Button>
            </CardContent>
          </Card>
        )}
      </main>
    </div>
  );
}
