package controller

// +kubebuilder:rbac:groups=supabase.k8s.icb4dc0.de,resources=cores,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=supabase.k8s.icb4dc0.de,resources=cores/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=supabase.k8s.icb4dc0.de,resources=cores/finalizers,verbs=update
// +kubebuilder:rbac:groups=supabase.k8s.icb4dc0.de,resources=apigateways,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=supabase.k8s.icb4dc0.de,resources=apigateways/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=supabase.k8s.icb4dc0.de,resources=apigateways/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=secrets;configmaps;services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=events,verbs=create
