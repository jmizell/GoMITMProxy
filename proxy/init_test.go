// Copyright 2019 The Jeremy Mizell. All rights reserved.
// Use of this source code is governed by a GPLv3 license that can be found in the LICENSE file.

package proxy

import "github.com/jmizell/GoMITMProxy/proxy/log"

func init() {
	log.DefaultLogger.Level = log.WARNING
}
