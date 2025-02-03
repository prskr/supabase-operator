/*
Copyright 2025 Peter Kurfer.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package health

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
)

var ErrCertExpired = errors.New("certificate expired")

func CertValidCheck(cert tls.Certificate) healthz.Checker {
	return func(req *http.Request) error {
		if cert.Leaf.NotAfter.Before(time.Now()) {
			err := fmt.Errorf("%w: %s", ErrCertExpired, cert.Leaf.Subject.CommonName)
			ctrl.Log.Error(err, "certificate expired", "commonName", cert.Leaf.Subject.CommonName, "not_after", cert.Leaf.NotAfter)
			return err
		}

		return nil
	}
}
