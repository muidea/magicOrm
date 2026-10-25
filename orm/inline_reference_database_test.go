package orm

import (
	"context"
	"crypto/rand"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
)

// 必须同时选择隔离授权标记与显式数据库；只通过 ORM 管理随机前缀的测试表。
// PostgreSQL 使用默认 build tag，MySQL 使用 -tags=mysql；不创建或修改共享 Schema。
func TestInlineReferenceIsolatedDatabaseConcurrency(t *testing.T) {
	if testing.Short() || os.Getenv("MAGICORM_INLINE_REFERENCE_TEST_ALLOW") != "isolated" {
		t.Skip("需要明确授权的隔离数据库及 MAGICORM_INLINE_REFERENCE_TEST_ALLOW=isolated")
	}
	server, name, schema := os.Getenv("MAGICORM_INLINE_REFERENCE_TEST_SERVER"), os.Getenv("MAGICORM_INLINE_REFERENCE_TEST_DATABASE"), os.Getenv("MAGICORM_INLINE_REFERENCE_TEST_SCHEMA")
	username, password := os.Getenv("MAGICORM_INLINE_REFERENCE_TEST_USER"), os.Getenv("MAGICORM_INLINE_REFERENCE_TEST_PASSWORD")
	if server == "" || !strings.HasPrefix(name, "magicorm_test_") || schema == "" || schema == "public" || username == "" || password == "" {
		t.Fatal("隔离数据库配置不完整")
	}
	cfg := NewConfig(server, name, schema, username, password)
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	for _, kind := range []string{"local", "remote"} {
		t.Run(kind, func(t *testing.T) {
			prefix := "inline_" + strings.ToLower(rand.Text())
			open := func(p provider.Provider, scope string) *impl {
				t.Helper()
				h, err := NewOrm(p, cfg, scope)
				if err != nil {
					t.Fatal("隔离 ORM 连接失败", err.Code)
				}
				h.(*impl).context = ctx
				return h.(*impl)
			}
			p, sample := inlineReferenceFixture(t, kind, &inlineReferenceSample{})
			target, err := p.GetTypeModel(sample.GetField("target").GetType())
			if err != nil {
				t.Fatal(err)
			}
			h := open(p, prefix)
			defer h.Release()
			if err := h.Create(target); err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := h.Drop(target); err != nil {
					t.Error(err)
				}
			}()
			if err := h.Create(sample); err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := h.Drop(sample); err != nil {
					t.Error(err)
				}
			}()
			if report, err := h.InspectSchema(sample, true); err != nil || report == nil || !report.Matches || len(report.Tables) != 1 {
				t.Fatal("物理主表引用结构不一致", report, err)
			}
			targetIDs := []int64{}
			for range 2 {
				value := target.Copy(models.MetaView)
				created, err := h.Insert(value)
				if err != nil {
					t.Fatal(err)
				}
				targetIDs = append(targetIDs, created.GetPrimaryField().GetValue().Get().(int64))
			}
			targetID := targetIDs[0]
			type outcome struct {
				id  any
				err *cd.Error
			}
			results := make(chan outcome, 2)
			start := make(chan struct{})
			var wg sync.WaitGroup
			for range 2 {
				writerProvider, value := inlineReferenceFixture(t, kind, &inlineReferenceSample{Stamp: 7, Target: &inlineReferenceTarget{ID: targetID}})
				writer := open(writerProvider, prefix)
				wg.Add(1)
				go func() {
					defer wg.Done()
					defer writer.Release()
					<-start
					created, err := writer.Insert(value)
					result := outcome{err: err}
					if created != nil {
						result.id = created.GetPrimaryField().GetValue().Get()
					}
					results <- result
				}()
			}
			close(start)
			wg.Wait()
			close(results)
			success, duplicate := 0, 0
			for result := range results {
				if result.err == nil && result.id != nil {
					success++
				} else if result.err != nil && result.err.Code == cd.Duplicated {
					duplicate++
				} else {
					t.Fatal("并发写产生非预期结果", result.err)
				}
			}
			if success != 1 || duplicate != 1 {
				t.Fatal("真实数据库未保证单条成功", success, duplicate)
			}
			filter, err := p.GetModelFilter(sample)
			if err != nil {
				t.Fatal(err)
			}
			if err := filter.Equal("stamp", int64(7)); err != nil {
				t.Fatal(err)
			}
			if count, err := h.Count(filter); err != nil || count != 1 {
				t.Fatal("并发后独立计数不一致", count, err)
			}
			_, different := inlineReferenceFixture(t, kind, &inlineReferenceSample{Stamp: 7, Target: &inlineReferenceTarget{ID: targetIDs[1]}})
			if _, err := h.Insert(different); err != nil {
				t.Fatal("不同目标同采样时间被拒绝", err)
			}
			// 独立实例表使用相同自然键，不能被另一实例的唯一约束误伤。
			other := open(p, prefix+"_b")
			defer other.Release()
			if err := other.Create(target); err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := other.Drop(target); err != nil {
					t.Error(err)
				}
			}()
			if err := other.Create(sample); err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := other.Drop(sample); err != nil {
					t.Error(err)
				}
			}()
			otherTarget, err := other.Insert(target.Copy(models.MetaView))
			if err != nil {
				t.Fatal(err)
			}
			_, value := inlineReferenceFixture(t, kind, &inlineReferenceSample{Stamp: 7, Target: &inlineReferenceTarget{ID: otherTarget.GetPrimaryField().GetValue().Get().(int64)}})
			if _, err := other.Insert(value); err != nil {
				t.Fatal("独立实例同自然键被拒绝", err)
			}
		})
	}
}
