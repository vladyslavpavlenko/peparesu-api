package main

import (
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/vladyslavpavlenko/peparesu/config"
	"github.com/vladyslavpavlenko/peparesu/internal/handlers"
	"github.com/vladyslavpavlenko/peparesu/internal/models"
	"github.com/vladyslavpavlenko/peparesu/internal/render"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
	"time"
)

func setup(app *config.AppConfig) error {
	// Get environment variables
	env, err := loadEvnVariables()
	if err != nil {
		return err
	}

	app.Env = env

	// Connect to the database and run migrations
	db, err := connectToPostgresAndMigrate(env)
	if err != nil {
		return err
	}

	app.DB = db

	// Run database migrations
	err = runDatabaseMigrations(db)
	if err != nil {
		return err
	}

	repo := handlers.NewRepo(app)
	handlers.NewHandlers(repo)
	render.NewRenderer(app)

	return nil
}

// loadEvnVariables loads variables from the .env file.
func loadEvnVariables() (*config.EnvVariables, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("error getting environment variables: %v", err)
	}

	postgresHost := os.Getenv("POSTGRES_HOST")
	postgresUser := os.Getenv("POSTGRES_USER")
	postgresPass := os.Getenv("POSTGRES_PASS")
	postgresDBName := os.Getenv("POSTGRES_DBNAME")
	jwtSecret := os.Getenv("JWT_SECRET")

	return &config.EnvVariables{
		PostgresHost:   postgresHost,
		PostgresUser:   postgresUser,
		PostgresPass:   postgresPass,
		PostgresDBName: postgresDBName,
		JWTSecret:      jwtSecret,
	}, nil
}

// connectToPostgresAndMigrate initializes a PostgreSQL db session and runs GORM migrations.
func connectToPostgresAndMigrate(env *config.EnvVariables) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s dbname=%s password=%s sslmode=disable",
		env.PostgresHost, env.PostgresUser, env.PostgresDBName, env.PostgresPass)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("could not connect: ", err)
	}

	return db, nil
}

func runDatabaseMigrations(db *gorm.DB) error {
	// create tables
	err := db.AutoMigrate(&models.UserType{})
	if err != nil {
		return err
	}

	err = db.AutoMigrate(&models.User{})
	if err != nil {
		return err
	}

	err = db.AutoMigrate(&models.Restaurant{})
	if err != nil {
		return err
	}

	err = db.AutoMigrate(&models.Menu{})
	if err != nil {
		return err
	}

	err = db.AutoMigrate(&models.MenuItem{})
	if err != nil {
		return err
	}

	// populate tables with initial data
	err = createInitialUserTypes(db)
	if err != nil {
		return errors.New(fmt.Sprint("error creating initial user types:", err))
	}

	err = createInitialUsers(db)
	if err != nil {
		return errors.New(fmt.Sprint("error creating initial users:", err))
	}

	err = createInitialRestaurants(db)
	if err != nil {
		return errors.New(fmt.Sprint("error creating initial restaurants:", err))
	}

	err = createInitialMenus(db)
	if err != nil {
		return errors.New(fmt.Sprint("error creating initial menus:", err))
	}

	err = createInitialMenuItems(db)
	if err != nil {
		return errors.New(fmt.Sprint("error creating initial menus:", err))
	}

	return nil
}

// createInitialUserTypes creates initial user types in the `user_types` table.
func createInitialUserTypes(db *gorm.DB) error {
	var count int64

	if err := db.Model(&models.UserType{}).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	initialData := []models.UserType{
		{Title: "User"},
		{Title: "Admin"},
	}

	if err := db.Create(&initialData).Error; err != nil {
		return err
	}

	return nil
}

// createInitialUserTypes creates initial users in the `users` table.
func createInitialUsers(db *gorm.DB) error {
	var count int64

	if err := db.Model(&models.User{}).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	initialData := []models.User{
		{
			FirstName:  "Владислав",
			LastName:   "Павленко",
			Email:      "mail@peparesu.com",
			Password:   "password",
			UserTypeID: 2, // Admin
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
		{
			FirstName:  "Алекс",
			LastName:   "Купер",
			Email:      "alex@cooper.com",
			Password:   "password",
			UserTypeID: 1, // User
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
		{
			FirstName:  "Михайло",
			LastName:   "Кацурін",
			Email:      "misha@katsurin.com",
			Password:   "password",
			UserTypeID: 1, // User
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
	}

	if err := db.Create(&initialData).Error; err != nil {
		return err
	}

	return nil
}

// createInitialRestaurants creates initial restaurants in the `restaurants` table.
func createInitialRestaurants(db *gorm.DB) error {
	var count int64

	if err := db.Model(&models.Restaurant{}).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	initialData := []models.Restaurant{
		{
			OwnerID:     1,
			Title:       "Молодість",
			Type:        "Кафе-бар",
			Description: "Гастро-відпустка у минуле! Обідаємо як у бабусі, згадуємо молодість і танцюємо під знайомі хіти вечорами.",
			Address:     "вулиця Князів Острозьких, 8, Київ, Україна, 02000",
			Phone:       "+380977041319",
		},
		{
			OwnerID: 2,
			Title:   "Японський привіт",
			Type:    "Ресторан",
			Address: "вулиця Рейтарська, 15, Київ, Україна, 02000",
			Phone:   "+380968007877",
		},
		{
			OwnerID:     2,
			Title:       "Тайський привіт",
			Type:        "Ресторан",
			Description: "Тайський Привіт — гастрономічний телепорт у Таїланд в центрі Києва. Ви знайдете тут все, що знали, і чого не знали про тайську кухню, а в інтер'єрі побачите справжній тайський антикваріат. Тайський Привіт — це чесна тайська їжа, дикий чай з джунглів, натуральне вино і справжній тайський масаж, який вам зроблять прямо в ресторані.",
			Address:     "Чеховський провулок, 2, Київ, Україна, 02000",
			Phone:       "+380508455505",
		},
	}

	if err := db.Create(&initialData).Error; err != nil {
		return err
	}

	return nil
}

// createInitialMenus creates initial menus in the `menus` table.
func createInitialMenus(db *gorm.DB) error {
	var count int64

	if err := db.Model(&models.Menu{}).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	initialData := []models.Menu{
		{
			RestaurantID: 1,
			Title:        "Сети закусок",
		},
		{
			RestaurantID: 1,
			Title:        "К-а-р-т-о-ш-к-а",
		},
		{
			RestaurantID: 1,
			Title:        "З тіста",
		},
		{
			RestaurantID: 1,
			Title:        "Коктейлі",
		},
		{
			RestaurantID: 2,
			Title:        "нігірі",
		},
		{
			RestaurantID: 2,
			Title:        "сети",
		},
		{
			RestaurantID: 2,
			Title:        "бар",
		},
		{
			RestaurantID: 3,
			Title:        "Том (супи)",
		},
		{
			RestaurantID: 3,
			Title:        "Напої",
		},
	}

	if err := db.Create(&initialData).Error; err != nil {
		return err
	}

	return nil
}

// createInitialMenuItems creates initial menu items in the `menu_items` table.
func createInitialMenuItems(db *gorm.DB) error {
	var count int64

	if err := db.Model(&models.MenuItem{}).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	initialData := []models.MenuItem{
		{
			MenuID:      1,
			Picture:     "https://www.scythia.vn.ua/wp-content/uploads/2021/12/IMG_0919-2.jpg",
			Title:       "Сет Бутербродний",
			Description: "Найсмачніші бутери: з сирокопченою ковбасою та вершковим маслом, з лимонним маслом і червоною ікрою.",
			PriceUAH:    910,
		},
		{
			MenuID:      1,
			Picture:     "https://cdn-media.choiceqr.com/prod-eat-molodist/menu/mwPKHII-CxChCuC-sHSGnny.jpeg.webp",
			Title:       "Сет із намазками",
			Description: "Паштет, зелене сало, еврейська намазка, ікра з баклажанів, форшмак, лечо з перців.",
			PriceUAH:    320,
		},
		{
			MenuID:      2,
			Picture:     "https://cdn-media.choiceqr.com/prod-eat-molodist/menu/cSNKvUJ-sDxhZqT-ciFHWWr.webp",
			Title:       "Сирна картошка",
			Description: "Картоплю смажимо на суміші топленого жиру зі спеціями. Подаємо з насиченим сирним соусом та міксом трьох видів сиру.",
			PriceUAH:    285,
		},
		{
			MenuID:      2,
			Picture:     "https://cdn-media.choiceqr.com/prod-eat-molodist/menu/xGHFnjq-OxHhgwE-wZUtVZv.webp",
			Title:       "Картошка з мортаделою та яйцем",
			Description: "Картоплю смажимо на суміші топленого жиру зі спеціями. Подаємо з насиченим сирним соусом, мортаделою обсмаженою.",
			PriceUAH:    300,
		},
		{
			MenuID:      2,
			Picture:     "https://cdn-media.choiceqr.com/prod-eat-molodist/menu/xNmYCYk-GttRgHQ-ckoZGNC.webp",
			Title:       "Смажена картопля зі шкварками",
			Description: "Картоплю смажимо на суміші топленого жиру зі спеціями. Подаємо зі шкварочками та зеленню.",
			PriceUAH:    320,
		},
		{
			MenuID:      3,
			Picture:     "https://cdn-media.choiceqr.com/prod-eat-molodist/menu/CgEMnWx-bIDPIUX-DjOaGyo.jpeg.webp",
			Title:       "Пельмені на всю стипендію",
			Description: "З куркою.",
			PriceUAH:    170,
		},
		{
			MenuID:      3,
			Picture:     "https://cdn-media.choiceqr.com/prod-eat-molodist/menu/CgEMnWx-bIDPIUX-DjOaGyo.jpeg.webp",
			Title:       "Пельмені на всю стипендію",
			Description: "Зі свининою.",
			PriceUAH:    175,
		},
		{
			MenuID:      3,
			Picture:     "https://cdn-media.choiceqr.com/prod-eat-molodist/menu/fGHLXng-UACmFmP-VCvefkd.jpeg.webp",
			Title:       "Пельмені смажені",
			Description: "Подаємо з вершково-грибним соусом та сиром моцарелла.",
			PriceUAH:    235,
		},
		{
			MenuID:      4,
			Picture:     "https://cdn-media.choiceqr.com/prod-eat-molodist/menu/ZtpkXlH-vmkUwcF-FZGeQLl.webp",
			Title:       "Піна Колада",
			Description: "CAPTAIN MORGAN TIKI, CAPTAIN MORGAN WHITE, PINEAPPLE JUICE, SOUR-CREAM",
			PriceUAH:    220,
		},
		{
			MenuID:      4,
			Picture:     "https://cdn-media.choiceqr.com/prod-eat-molodist/menu/CwabAgL-RMkVzke-pNWNQjo.webp",
			Title:       "Big Lebowski",
			Description: "Сoffee liqueur, Vodka Koskenkorva, sour cream",
			PriceUAH:    220,
		},
		{
			MenuID:      5,
			Picture:     "https://cdn-media.choiceqr.com/prod-eat-japanhi-privet-delivery/menu/GsQBadX-zxIpnee-LubUKXH.jpeg.webp",
			Title:       "нігірі з лососем і домашнім унагі",
			Description: "з цибулею шніт, томатним айолі та кунжутом юзу\n",
			PriceUAH:    115,
		},
		{
			MenuID:      5,
			Picture:     "https://cdn-media.choiceqr.com/prod-eat-japanhi-privet-delivery/menu/TUseGDG-DkVPTre-KddHHmC.jpeg.webp",
			Title:       "нігірі з гребінцем",
			Description: "з соусом місо",
			PriceUAH:    210,
		},
		{
			MenuID:      5,
			Picture:     "https://cdn-media.choiceqr.com/prod-eat-japanhi-privet-delivery/menu/KHCXibN-GMUtxzV-MQwzlPD.jpeg.webp",
			Title:       "нігірі з лангустином і сальсою манго",
			Description: "з соусом вінегрет юзу",
			PriceUAH:    120,
		},
		{
			MenuID:      5,
			Picture:     "https://cdn-media.choiceqr.com/prod-eat-japanhi-privet-delivery/menu/NOHzvNq-vCjzsJV-VACAfGX.jpeg.webp",
			Title:       "нігірі з тунцем",
			Description: "з цибулею шніт, кунжутом кімчі та домашнім унагі",
			PriceUAH:    115,
		},
		{
			MenuID:      6,
			Picture:     "https://cdn-media.choiceqr.com/prod-eat-japanhi-privet-delivery/menu/XtCiBbv-RFHuGYb-NRvnFGn.jpeg.webp",
			Title:       "сет 1",
			Description: "рол з лососем або вугром, філадельфією, огірком і унагі, футомакі з тунцем, лососем, шиітаке, огірком і соусом джпн хай.",
			PriceUAH:    1360,
		},
		{
			MenuID:      6,
			Picture:     "https://cdn-media.choiceqr.com/prod-eat-japanhi-privet-delivery/menu/UIhHFpx-KfLIHJz-ViTnDMu.jpeg.webp",
			Title:       "сет 2",
			Description: "рол з лососем або вугром, філадельфією, огірком і унагі, рол з лангустином, лососем татакі, філадельфією, кисло-солодким соусом і трюфельним айолі.",
			PriceUAH:    1840,
		},
		{
			MenuID:      7,
			Picture:     "http://localhost:8080/api/v1/storage/images/menuitem-default.jpeg",
			Title:       "сенча",
			Description: "Чай з м'яким свіжим ароматом та солодким присмаком. Чудово тамує спрагу і наповнює енергією.",
			PriceUAH:    190,
		},
		{
			MenuID:   7,
			Picture:  "http://localhost:8080/api/v1/storage/images/menuitem-default.jpeg",
			Title:    "кабусеча генмайча",
			PriceUAH: 170,
		},
		{
			MenuID:      8,
			Picture:     "https://cdn-media.choiceqr.com/prod-eat-thailandhi/menu/RjXrcjv-EtdAzby-CkwIYBT.jpeg.webp",
			Title:       "Спарвжній Том Ям",
			Description: "кисло-гострий суп з креветками, кальмарами, лемонграсом, галангалом, соком лайма, зеленню та грибами ерінгами",
			PriceUAH:    380,
		},
		{
			MenuID:      8,
			Picture:     "https://cdn-media.choiceqr.com/prod-eat-thailandhi/menu/wIfWSja-KIHPZCf-lGRDqII.jpeg.webp",
			Title:       "Туристичний Том Ям",
			Description: "кисло-гострий суп з кокосовим молоком, креветками, кальмарами, лемонграсом, галангалом, соком лайма, зеленню та ерінгами",
			PriceUAH:    440,
		},
		{
			MenuID:      9,
			Picture:     "https://cdn-media.choiceqr.com/prod-eat-thailandhi/menu/gbSBllF-elDedEb-wseYgcv.jpeg.webp",
			Title:       "Ча Єн",
			Description: "чорний цейлонський чай з букетом східних спецій і згущеним молоком",
			PriceUAH:    110,
		},
		{
			MenuID:   9,
			Picture:  "https://cdn-media.choiceqr.com/prod-eat-thailandhi/menu/jYCfmUl-bHMvVNY-DbjEckS.jpeg.webp",
			Title:    "Манго-маракуя-матча",
			PriceUAH: 150,
		},
	}

	if err := db.Create(&initialData).Error; err != nil {
		return err
	}

	return nil
}
