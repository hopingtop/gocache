package gocache

import (
	"os"
	"testing"
	"time"
)

var dir, _ = os.Getwd()
var cfg = Config{
	MaxSize:       1000,
	CleanInterval: time.Millisecond * 500,
	AutoSave:      false,
	SaveType:      SaveAllKeysMode,
	Filename:      dir + "/cache.back",
}
var cache, _ = NewCache(cfg)

func TestGetAndSet(t *testing.T) {
	cache.Set("test", "123")

	v, ok := cache.Get("test")
	if !ok {
		t.Fail()
	}

	if v.(string) != "123" {
		t.Fatal(v.(string))
	}

	if err := cache.SetExpire("testExpire", "testExpire", time.Minute); err != nil {
		t.Fatal(err)
	}

	if err := cache.Set("testExpire", "testExpire"); err != nil {
		t.Fatal(err)
	}
	_, e, ok := cache.GetWithExpire("testExpire")
	if !ok || e.Nanoseconds() > 0 {
		t.Fail()
	}

	if err := cache.SetExpire("testExpire", "testExpire", time.Microsecond*100); err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Microsecond * 150)
	_, ok = cache.Get("testExpire")
	if ok {
		t.Fail()
	}
	if err := cache.SetExpire("testExpire", "testExpire", time.Microsecond*100); err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Microsecond * 150)
	_, e, ok = cache.GetWithExpire("testExpire")
	if ok || e.Nanoseconds() > 0 {
		t.Fail()
	}

	if err := cache.Add("test", 1, 0); err == nil {
		t.Fail()
	}
}

func TestExpireClean(t *testing.T) {
	cache.SetExpire("expire", "expire", time.Millisecond*200)
	v, ok := cache.Get("expire")
	if !ok {
		t.Fail()
	}

	if v.(string) != "expire" {
		t.Fatal(v.(string))
	}
	size := cache.Size()

	time.Sleep(time.Millisecond * 800)
	v, ok = cache.Get("expire")
	if ok {
		t.Fail()
	}
	if cache.Size()+1 != size {
		t.Fail()
	}
}

func TestAutoSaveAndLoad(t *testing.T) {
	cfg.SaveType = SaveAllKeysMode
	cfg.AutoSave = true

	c, err := NewCache(cfg)
	if err != nil {
		t.Fatal(err)
	}

	if err := c.Set("autoSaveAllKeysMode", "SaveAllKeysMode"); err != nil {
		t.Fatal(err)
	}
	if err := c.SetExpire("SaveAllKeysMode", "SaveAllKeysMode", time.Minute); err != nil {
		t.Fatal(err)
	}

	// 模拟程序正常退出 close
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}

	// 重置缓存, 自动加载已经保存的文件
	c, err = NewCache(cfg)

	v, ok := c.Get("autoSaveAllKeysMode")
	if !ok || v.(string) != "SaveAllKeysMode" {
		t.Fatal("load data loss")
	}

	v, ok = c.Get("SaveAllKeysMode")
	if !ok || v.(string) != "SaveAllKeysMode" {
		t.Fatal("load data loss")
	}
}

func TestSaveAndLoad(t *testing.T) {
	cfg.SaveType = SaveAllKeysMode
	c, err := NewCache(cfg)
	if err != nil {
		t.Fatal(err)
	}

	if err := c.Set("SaveAllKeysMode", "SaveAllKeysMode"); err != nil {
		t.Fatal(err)
	}
	c.SaveDisk(cfg.Filename, cfg.SaveType)

	if err := c.SaveDisk("", cfg.SaveType); err == nil {
		t.Fail()
	}

	cfg.AutoSave = false
	c.Close()

	// 重置缓存
	c, err = NewCache(cfg)
	if err := c.LoadDisk(cfg.Filename); err != nil {
		t.Fatal(err)
	}

	v, ok := c.Get("SaveAllKeysMode")
	if !ok || v.(string) != "SaveAllKeysMode" {
		t.Fatal("load data loss")
	}
}

func TestAutoSaveAndLoadMode(t *testing.T) {
	cfg.SaveType = SaveExpireKeysMode
	cfg.AutoSave = true

	c, err := NewCache(cfg)
	if err != nil {
		t.Fatal(err)
	}

	if err := c.Set("testSaveMode", "SaveAllKeysMode"); err != nil {
		t.Fatal(err)
	}

	if err := c.SetExpire("SaveExpireKeysMode", "SaveExpireKeysMode", time.Minute); err != nil {
		t.Fatal(err)
	}

	// 模拟程序正常退出 close
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}

	// 重置缓存, 自动加载已经保存的文件
	c, err = NewCache(cfg)

	v, ok := c.Get("SaveExpireKeysMode")
	if !ok || v.(string) != "SaveExpireKeysMode" {
		t.Fatal("SaveExpireKeysMode load data loss")
	}

	v, ok = c.Get("testSaveMode")
	if ok {
		t.Fatal("save data mode err")
	}

	// 保存永久 key
	cfg.SaveType = SaveNoExpireKeysMode

	c, err = NewCache(cfg)
	if err != nil {
		t.Fatal(err)
	}

	if err := c.Set("testSaveMode", "SaveAllKeysMode"); err != nil {
		t.Fatal(err)
	}

	if err := c.SetExpire("SaveExpireKeysMode", "SaveExpireKeysMode", time.Minute); err != nil {
		t.Fatal(err)
	}

	// 模拟程序正常退出 close
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}

	// 重置缓存, 自动加载已经保存的文件
	c, err = NewCache(cfg)

	v, ok = c.Get("testSaveMode")
	if !ok || v.(string) != "SaveAllKeysMode" {
		t.Fatal("SaveAllKeysMode load data loss")
	}

	v, ok = c.Get("SaveExpireKeysMode")
	if ok {
		t.Fatal("save data mode err")
	}
}

func TestCacheFunc(t *testing.T) {
	cfg.CleanInterval = time.Millisecond * 500
	cfg.MaxSize = 5
	cfg.OverSizeClearMode = NoEvictionMode
	c, _ := NewCache(cfg)
	c.Flush()
	c.Set("t1", 1)
	c.Set("t2", 2)
	c.Set("t3", 3)
	c.Set("t4", 4)
	c.SetExpire("t5", 5, time.Second)
	if err := c.SetExpire("t6", 6, time.Second); err == nil {
		t.Fatal("size limit fatal")
	}

	if c.Size() != 5 {
		t.Fatal("size loss")
	}

	time.Sleep(time.Second * 2)
	if c.Size() != 4 {
		t.Fatal("expired failure", c.Size())
	}

	if err := c.Add("t5", 5, time.Second*3); err != nil {
		t.Fatal(err)
	}

	if err := c.Add("t5", 5, 0); err == nil {
		t.Fatal("add failure")
	}

	c.Range(func(k, v interface{}) bool {
		if k.(string) == "t2" {
			return false
		}
		return true
	})

	v, e, ok := c.GetWithExpire("t5")
	if !ok {
		t.Fatal("Add err not found")
	}
	if v.(int) != 5 || e.Nanoseconds() <= 0 {
		t.Fail()
	}

}

func TestOverSizeVolatileRandomModeDeleteKey(t *testing.T) {
	cfg.MaxSize = 2
	cfg.OverSizeClearMode = VolatileRandomMode
	cfg.CleanInterval = 0
	c, _ := NewCache(cfg)
	c.Flush()
	if err := c.Set("t1", 1); err != nil {
		t.Fatal(err)
	}
	if err := c.Set("t2", 1); err != nil {
		t.Fatal(err)
	}
	if err := c.Set("t3", 1); err == nil {
		t.Fatal("t3 error")
	}
	c.Delete("t2")
	if err := c.SetExpire("t2", 1, time.Minute); err != nil {
		t.Fatal(err)
	}
	if err := c.Set("t3", 1); err != nil {
		t.Fatal("t3 error")
	}
	if err := c.SetExpire("t4", 1, time.Minute); err == nil {
		t.Fatal("t4 error")
	}
	if c.Size() > 2 {
		t.Fatal("size err", c.Size())
	}
}

func TestOverSizeAllKeysRandomModeDeleteKey(t *testing.T) {
	cfg.MaxSize = 2
	cfg.OverSizeClearMode = AllKeysRandomMode
	cfg.CleanInterval = 0
	c, _ := NewCache(cfg)
	c.Flush()
	if err := c.Set("t1", 1); err != nil {
		t.Fatal(err)
	}
	if err := c.Set("t2", 1); err != nil {
		t.Fatal(err)
	}
	if err := c.Set("t3", 1); err != nil {
		t.Fatal(err)
	}
	if c.Size() > 2 {
		t.Fatal("size err", c.Size())
	}
}
