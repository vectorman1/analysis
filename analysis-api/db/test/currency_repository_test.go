package test_test

import (
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/vectorman1/analysis/analysis-api/common"
	"github.com/vectorman1/analysis/analysis-api/db"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type CurrencyServiceSuite struct {
	repo *db.CurrencyRepository
	mock sqlmock.Sqlmock
}

var _ = Describe("CurrencyService", func() {
	var suite CurrencyServiceSuite
	var mdb *sql.DB

	BeforeEach(func() {
		appConfig, appCErr := common.GetConfig()
		Expect(appCErr).ShouldNot(HaveOccurred())

		suite = CurrencyServiceSuite{}

		var mockErr error
		mdb, suite.mock, mockErr = sqlmock.New()
		Expect(mockErr).ShouldNot(HaveOccurred())

		d := mysql.New(mysql.Config{
			DSN:                       appConfig.PostgreSQLConnString,
			DriverName:                "mysql",
			Conn:                      mdb,
			SkipInitializeWithVersion: true,
		})

		gdb, gdbErr := gorm.Open(d, &gorm.Config{})
		Expect(gdbErr).ShouldNot(HaveOccurred())

		suite.repo = db.NewCurrencyRepository(gdb)
	})
	AfterEach(func() {
		err := suite.mock.ExpectationsWereMet()
		Expect(err).ShouldNot(HaveOccurred())
		suite.mock = nil
	})

	Context("get by code", func() {
		It("empty by invalid code", func() {
			r, err := suite.repo.GetByCode("ZXCSD")
			Expect(err).Should(HaveOccurred())
			Expect(r).Should(BeNil())
		})
	})
})
