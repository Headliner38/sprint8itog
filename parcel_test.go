package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // настройте подключение к БД
	require.NoError(t, err, "error")
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	pkgId, err := store.Add(parcel)
	require.NoError(t, err, "error")
	require.NotEmpty(t, pkgId, "id is empty")

	// get
	// получите только что добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
	pkg, err := store.Get(pkgId)
	require.NoError(t, err, "error")

	parcel.Number = pkg.Number // надеюсь не что то позорное сделал <3
	assert.Equal(t, parcel, pkg, "values not equal")
	//assert.Equal(t, parcel.CreatedAt, pkg.CreatedAt, "values not equal")
	//assert.Equal(t, parcel.Status, pkg.Status, "values not equal")
	//assert.Equal(t, parcel.Client, pkg.Client, "values not equal")

	// delete
	// удалите добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что посылку больше нельзя получить из БД

	err = store.Delete(pkgId)
	require.NoError(t, err, "error")

	_, err = store.Get(pkgId)
	require.ErrorIs(t, err, sql.ErrNoRows, "something wrong...")
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // настройте подключение к БД
	require.NoError(t, err, "error")
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	pkgId, err := store.Add(parcel)
	require.NoError(t, err, "error")
	require.NotEmpty(t, pkgId, "id is empty")

	// set address
	// обновите адрес, убедитесь в отсутствии ошибки
	newAddress := "new test address"
	err = store.SetAddress(pkgId, newAddress)
	require.NoError(t, err, "error")

	// check
	// получите добавленную посылку и убедитесь, что адрес обновился
	pkg, err := store.Get(pkgId)
	require.NoError(t, err, "error")
	assert.Equal(t, newAddress, pkg.Address, "the value has not been updated")
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // настройте подключение к БД
	require.NoError(t, err, "error")
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	pkgId, err := store.Add(parcel)
	require.NoError(t, err, "error")
	require.NotEmpty(t, pkgId, "id is empty")

	// set status
	// обновите статус, убедитесь в отсутствии ошибки
	nStatus := ParcelStatusRegistered
	err = store.SetStatus(pkgId, nStatus)
	require.NoError(t, err, "error")

	// check
	// получите добавленную посылку и убедитесь, что статус обновился
	pkg, err := store.Get(pkgId)
	require.NoError(t, err, "error")
	assert.Equal(t, nStatus, pkg.Status, "status values not equal")
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // настройте подключение к БД
	require.NoError(t, err, "error")
	defer db.Close()

	store := NewParcelStore(db)
	//parcel := getTestParcel()

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		pkgId, err := store.Add(parcels[i]) // добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
		require.NoError(t, err, "error")
		require.NotEmpty(t, pkgId, "")

		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = pkgId

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[pkgId] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client) // получите список посылок по идентификатору клиента, сохранённого в переменной client
	// убедитесь в отсутствии ошибки
	// убедитесь, что количество полученных посылок совпадает с количеством добавленных
	require.NoError(t, err, "error")
	require.Equal(t, len(parcels), len(storedParcels), "the number of parcels received does not match the number of those added")

	// check
	for _, parcel := range storedParcels {
		// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		// убедитесь, что все посылки из storedParcels есть в parcelMap
		// убедитесь, что значения полей полученных посылок заполнены верно
		exp, err := parcelMap[parcel.Number]
		require.True(t, err, parcel.Number)
		require.Equal(t, exp, parcel)
	}
}
