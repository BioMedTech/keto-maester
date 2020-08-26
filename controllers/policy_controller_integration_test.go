package controllers_test

import (
	"context"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	keto1alpha1 "github.com/ory/keto-maester/api/v1alpha1"
	"github.com/ory/keto-maester/controllers"
	"github.com/ory/keto-maester/controllers/mocks"
	"github.com/ory/keto-maester/keto"
	. "github.com/stretchr/testify/mock"
	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"time"
)

const (
	timeout      = time.Second * 5
	tstNamespace = "default"
	tstSecret    = "testSecret"
)

var _ = Describe("Policy Controller", func() {

	Context("in a happy-path scenario", func() {

		Context("should call create Policy and", func() {

			It("succeed", func() {

				tstName := "test"
				expectedRequest := &reconcile.Request{NamespacedName: types.NamespacedName{Name: tstName, Namespace: tstNamespace}}

				s := scheme.Scheme
				err := keto1alpha1.AddToScheme(s)
				Expect(err).NotTo(HaveOccurred())

				err = apiv1.AddToScheme(s)
				Expect(err).NotTo(HaveOccurred())

				// Setup the Manager and Controller.  Wrap the Controller Reconcile function so it writes each request to a
				// channel when it is finished.
				mgr, err := manager.New(cfg, manager.Options{Scheme: s})
				Expect(err).NotTo(HaveOccurred())
				c := mgr.GetClient()

				mch := &mocks.KetoClient{}
				mch.On("GetPolicy", Anything).Return(nil, false, nil)
				mch.On("DeletePolicy", Anything).Return(nil)
				mch.On("ListPolicy", Anything).Return(nil, nil)
				mch.On("PostPolicy", AnythingOfType("*hydra.OAuth2ClientJSON")).Return(func(o *keto.PolicyJSON) *keto.PolicyJSON {
					return &keto.PolicyJSON{}
				}, func(o *keto.PolicyJSON) error {
					return nil
				})

				recFn, requests := SetupTestReconcile(getAPIReconciler(mgr, mch))

				Expect(add(mgr, recFn)).To(Succeed())

				//Start the manager and the controller
				stopMgr, mgrStopped := StartTestManager(mgr)

				instance := testInstance(tstName)
				err = c.Create(context.TODO(), instance)
				// The instance object may not be a valid object because it might be missing some required fields.
				// Please modify the instance object by adding required fields and then remove the following if statement.
				if apierrors.IsInvalid(err) {
					Fail(fmt.Sprintf("failed to create object, got an invalid object error: %v", err))
					return
				}
				Expect(err).NotTo(HaveOccurred())
				Eventually(requests, timeout).Should(Receive(Equal(*expectedRequest)))

				//Verify the created CR instance status
				var retrieved keto1alpha1.Policy
				ok := client.ObjectKey{Name: tstName, Namespace: tstNamespace}
				err = c.Get(context.TODO(), ok, &retrieved)
				Expect(err).NotTo(HaveOccurred())
				Expect(retrieved.Status.ReconciliationError.Description).To(BeEmpty())

				//delete instance
				c.Delete(context.TODO(), instance)

				//Ensure manager is stopped properly
				close(stopMgr)
				mgrStopped.Wait()
			})

		})
	})
})

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("api-gateway-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Api
	err = c.Watch(&source.Kind{Type: &keto1alpha1.Policy{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

func getAPIReconciler(mgr ctrl.Manager, mock controllers.KetoClient) reconcile.Reconciler {
	return &controllers.KetoPolicyReconciler{
		&controllers.Reconciler{
			Client:     mgr.GetClient(),
			Log:        ctrl.Log.WithName("controllers").WithName("OAuth2Client"),
			KetoClient: mock,
		},
	}
}

func testInstance(name string) *keto1alpha1.Policy {

	return &keto1alpha1.Policy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: tstNamespace,
		},
		Spec: keto1alpha1.PolicySpec{
			PatternMatching: "exact",
			Description:     "allow maria to get photos",
			Subjects:        []string{"users:maria"},
			Actions:         []string{"list"},
			Effect:          "allow",
			Resources:       []string{"resources:photos"},
			Conditions:      nil,
		}}
}
