package main

import (
	"fmt"
	"log"
	"project/config"
)

func main() {
	fmt.Println("🎮 Game Server Configuration Demo")
	fmt.Println("================================")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Display Game Configuration
	fmt.Printf("\n🎯 Game Settings:\n")
	fmt.Printf("  Name: %s\n", cfg.Game.Name)
	fmt.Printf("  Version: %s\n", cfg.Game.Version)
	fmt.Printf("  Max Players: %d\n", cfg.Game.MaxPlayers)
	fmt.Printf("  Difficulty: %s\n", cfg.Game.Difficulty)
	fmt.Printf("  PvP Enabled: %t\n", cfg.Game.PvpEnabled)

	// Display World Configuration
	fmt.Printf("\n🌍 World Settings:\n")
	fmt.Printf("  World Name: %s\n", cfg.Game.World.Name)
	fmt.Printf("  World Seed: %s\n", cfg.Game.World.Seed)
	fmt.Printf("  World Size: %s\n", cfg.Game.World.Size)
	fmt.Printf("  Weather Enabled: %t\n", cfg.Game.World.WeatherEnabled)
	fmt.Printf("  Day/Night Cycle: %t\n", cfg.Game.World.DayNightCycle)
	fmt.Printf("  Spawn Point: X=%d, Y=%d, Z=%d\n",
		cfg.Game.World.SpawnPoint.X,
		cfg.Game.World.SpawnPoint.Y,
		cfg.Game.World.SpawnPoint.Z)

	// Display Player Configuration
	fmt.Printf("\n👤 Player Settings:\n")
	fmt.Printf("  Starting Health: %d\n", cfg.Game.Player.StartingHealth)
	fmt.Printf("  Starting Money: %s\n", cfg.Game.Player.StartingMoney)
	fmt.Printf("  Max Inventory Slots: %d\n", cfg.Game.Player.MaxInventorySlots)
	fmt.Printf("  Respawn Time: %d seconds\n", cfg.Game.Player.RespawnTime)
	fmt.Printf("  Starter Kit: %v\n", cfg.Game.Player.StarterKit)

	// Display Web Server Configuration
	fmt.Printf("\n🌐 Web Server Settings:\n")
	fmt.Printf("  Host: %s\n", cfg.Web.Host)
	fmt.Printf("  Port: %s\n", cfg.Web.Port)
	fmt.Printf("  SSL Enabled: %s\n", cfg.Web.SslEnabled)
	fmt.Printf("  Admin Panel: %t\n", cfg.Web.AdminPanel)
	fmt.Printf("  API Rate Limit: %d\n", cfg.Web.Api.RateLimit)
	fmt.Printf("  API Timeout: %s\n", cfg.Web.Api.Timeout)
	fmt.Printf("  CORS Enabled: %t\n", cfg.Web.Api.CorsEnabled)

	// Display Database Configuration
	fmt.Printf("\n🗄️  Database Settings:\n")
	fmt.Printf("  Type: %s\n", cfg.Database.Type)
	fmt.Printf("  Connection: %s\n", cfg.Database.Connection)
	fmt.Printf("  Max Connections: %d\n", cfg.Database.Pool.MaxConnections)
	fmt.Printf("  Min Connections: %d\n", cfg.Database.Pool.MinConnections)
	fmt.Printf("  Auto Migrate: %s\n", cfg.Database.Migrations.AutoMigrate)

	// Display Features Configuration
	fmt.Printf("\n🎪 Features Settings:\n")
	fmt.Printf("  Chat Enabled: %t\n", cfg.Features.Chat.Enabled)
	fmt.Printf("  Max Message Length: %d\n", cfg.Features.Chat.MaxMessageLength)
	fmt.Printf("  Chat Channels: %v\n", cfg.Features.Chat.Channels)
	fmt.Printf("  Economy Tax Rate: %s\n", cfg.Features.Economy.TaxRate)
	fmt.Printf("  Daily Bonus: %d\n", cfg.Features.Economy.DailyBonus)
	fmt.Printf("  Boss Fights Enabled: %t\n", cfg.Features.Events.BossFights.Enabled)
	fmt.Printf("  Boss Fight Min Players: %d\n", cfg.Features.Events.BossFights.MinPlayers)

	// Display Cache Configuration
	fmt.Printf("\n💾 Cache Settings:\n")
	fmt.Printf("  Type: %s\n", cfg.Cache.Type)
	fmt.Printf("  Redis Host: %s\n", cfg.Cache.Redis.Host)
	fmt.Printf("  Redis Port: %s\n", cfg.Cache.Redis.Port)
	fmt.Printf("  Player Data TTL: %s\n", cfg.Cache.Ttl.PlayerData)
	fmt.Printf("  Leaderboards TTL: %s\n", cfg.Cache.Ttl.Leaderboards)

	// Display Security Configuration
	fmt.Printf("\n🔒 Security Settings:\n")
	fmt.Printf("  Anticheat Enabled: %t\n", cfg.Security.Anticheat.Enabled)
	fmt.Printf("  Anticheat Strict Mode: %s\n", cfg.Security.Anticheat.StrictMode)
	fmt.Printf("  Rate Limiting Enabled: %t\n", cfg.Security.RateLimiting.Enabled)
	fmt.Printf("  Requests Per Minute: %d\n", cfg.Security.RateLimiting.RequestsPerMinute)

	// Display Monitoring Configuration
	fmt.Printf("\n📊 Monitoring Settings:\n")
	fmt.Printf("  Metrics Enabled: %t\n", cfg.Monitoring.Metrics.Enabled)
	fmt.Printf("  Metrics Endpoint: %s\n", cfg.Monitoring.Metrics.Endpoint)
	fmt.Printf("  Log Level: %s\n", cfg.Monitoring.Logging.Level)
	fmt.Printf("  Log Format: %s\n", cfg.Monitoring.Logging.Format)
	fmt.Printf("  File Logging Enabled: %s\n", cfg.Monitoring.Logging.File.Enabled)

	fmt.Printf("\n✅ Configuration loaded successfully!\n")
	fmt.Printf("🚀 Server ready to start with these settings.\n")
}
