package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kyma-incubator/compass/components/instance-creator/internal/client/types"
	"github.com/kyma-incubator/compass/components/instance-creator/internal/client/types/tenantmapping"
	"github.com/kyma-incubator/compass/components/instance-creator/internal/handler"
	"github.com/kyma-incubator/compass/components/instance-creator/internal/handler/automock"
	persistenceautomock "github.com/kyma-incubator/compass/components/instance-creator/internal/persistence/automock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

const (
	assignOperation           = "assign"
	unassignOperation         = "unassign"
	inboundCommunicationKey   = "inboundCommunication"
	serviceInstancesKey       = "serviceInstances"
	serviceBindingKey         = "serviceBinding"
	serviceInstanceServiceKey = "service"
	serviceInstancePlanKey    = "plan"
	configurationKey          = "configuration"
	nameKey                   = "name"
	assignmentIDKey           = "assignment_id"
)

var (
	formationID                  = "formation-id"
	assignmentID                 = "assignment-id"
	region                       = "region"
	subaccount                   = "subaccount"
	serviceInstancesIDs          = []string{"instance-id-1", "instance-id-2"}
	serviceInstancesNames        = []string{"instance-name-1", "instance-name-2"}
	serviceInstancesBindingsIDs  = []string{"binding-id-1", "binding-id-2", "binding-id-3", "binding-id-4"}
	serviceInstanceBindingsNames = []string{"binding-name-1", "binding-name-2"}
	serviceOfferingIDs           = []string{"service-offering-id-1", "service-offering-id-2"}
	servicePlanIDs               = []string{"service-plan-id-1", "service-plan-id-2"}

	serviceInstanceID          = "instance-id"
	serviceInstanceName        = "instance-name"
	serviceInstanceBindingID   = "binding-id"
	serviceInstanceBindingName = "binding-name"
	serviceOfferingID          = "service-offering-id"
	servicePlanID              = "service-plan-id"
)

func Test_HandlerFunc(t *testing.T) {
	url := "https://target-url.com"
	apiPath := fmt.Sprintf("/")
	statusUrl := "localhost"

	testErr := errors.New("test error")

	emptyJSON := `{}`

	reqBodyFormatter := `{
	 "context": %s,
	 "receiverTenant": %s,
	 "assignedTenant": %s
	}`

	reqBodyContextFormatter := `{"uclFormationId": %q, "operation": %q}`
	reqBodyContextWithAssign := fmt.Sprintf(reqBodyContextFormatter, formationID, assignOperation)
	reqBodyContextWithUnassign := fmt.Sprintf(reqBodyContextFormatter, formationID, unassignOperation)

	assignedTenantFormatter := `{
		"uclAssignmentId": %q,
		"configuration": %s
	}`

	assignedTenantConfigurationWithGlobalInstancesWithoutJsonpaths := `{
	      "credentials": {
            "inboundCommunication": {
			  "serviceInstances": [
                {
                  "service": "procurement-service",
                  "plan": "apiaccess",
                  "parameters": {},
                  "serviceBinding": {
                    "parameters": {}
                  }
                },
                {
                  "service": "identity",
                  "plan": "application",
                  "parameters": {
                    "consumed-services": [],
                    "xsuaa-cross-consumption": true
                  },
                  "serviceBinding": {
                    "parameters": {
                      "credential-type": "X509_PROVIDED",
                      "certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----"
                    }
                  }
                }
              ],
			  "no-instances-auth-method": {
                "certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----"
			  },
			  "refering-global-instances-auth-method": {
                "certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----"
			  }
            }
          }
	    }`
	expectedResponseForGlobalInstances := `{"state":"READY","configuration":{"credentials":{"outboundCommunication":{"no-instances-auth-method":{"certificate":"-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----"},"refering-global-instances-auth-method":{"certificate":"-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----"}}}}}`

	assignedTenantConfigurationWithGlobalInstancesWithJsonpaths := `{
	     "credentials": {
	       "inboundCommunication": {
			  "serviceInstances": [
	           {
	             "service": "procurement-service",
	             "plan": "apiaccess",
	             "parameters": {},
	             "serviceBinding": {
	               "parameters": {
				      "service-instance-plan": "$.credentials.inboundCommunication.serviceInstances[0].plan"
			        }
	             }
	           },
	           {
	             "service": "identity",
	             "plan": "application",
	             "parameters": {
	               "consumed-services": [
	                 {
	                   "first-service-instance-service": "$.credentials.inboundCommunication.serviceInstances[0].service"
	                 }
	               ],
	               "xsuaa-cross-consumption": true
	             },
	             "serviceBinding": {
	               "parameters": {
	                 "credential-type": "X509_PROVIDED",
	                 "certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----",
				      "first-service-instance-plan": "$.credentials.inboundCommunication.serviceInstances[0].plan",
				      "service-instance-plan": "$.credentials.inboundCommunication.serviceInstances[1].plan"
	               }
	             }
	           }
	         ],
			  "no-instances-auth-method": {
	           "certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----"
			  },
			  "refering-global-instances-auth-method": {
				"certificate": "$.credentials.inboundCommunication.serviceInstances[1].serviceBinding.parameters.certificate"
			  }
	       }
	     }
	   }`
	substitutedAssignedTenantConfigurationWithGlobalInstancesWithJsonpaths := `{
	     "credentials": {
	       "inboundCommunication": {
			  "serviceInstances": [
	           {
	             "service": "procurement-service",
	             "plan": "apiaccess",
	             "parameters": {},
	             "serviceBinding": {
	               "parameters": {
				      "service-instance-plan": "apiaccess"
			        }
	             }
	           },
	           {
	             "service": "identity",
	             "plan": "application",
	             "parameters": {
	               "consumed-services": [
	                 {
	                   "first-service-instance-service": "procurement-service"
	                 }
	               ],
	               "xsuaa-cross-consumption": true
	             },
	             "serviceBinding": {
	               "parameters": {
	                 "credential-type": "X509_PROVIDED",
	                 "certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----",
				      "first-service-instance-plan": "apiaccess",
				      "service-instance-plan": "application"
	               }
	             }
	           }
	         ],
			  "no-instances-auth-method": {
	           "certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----"
			  },
			  "refering-global-instances-auth-method": {
				"certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----"
			  }
	       }
	     }
	   }`

	assignedTenantConfigurationWithLocalInstancesWithJsonpaths := `{
	     "credentials": {
	       "inboundCommunication": {
			  "auth_method_with_local_instances": {
        		  "tokenServiceUrl": "$.credentials.inboundCommunication.auth_method_with_local_instances.serviceInstances[1].serviceBinding.url",
        		  "clientId": "$.credentials.inboundCommunication.auth_method_with_local_instances.serviceInstances[1].serviceBinding.clientid",
        		  "certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----",
        		  "correlationIds": ["ASD"],
        		  "serviceInstances": [
        		    {
        		      "service": "service-test",
        		      "plan": "plan-test",
        		      "configuration": {},
        		      "serviceBinding": {
        		        "configuration": {}
        		      }
        		    },
        		    {
        		      "service": "service-test-2",
        		      "plan": "plan-test-2",
        		      "configuration": {
        		        "consumed-services": [
        		          {
        		            "first-service-instance-plan": "$.credentials.inboundCommunication.auth_method_with_local_instances.serviceInstances[0].plan"
        		          }
        		        ],
        		        "xsuaa-cross-consumption": true
        		      },
        		      "serviceBinding": {
						"url": "url",
						"clientid": "clientid",
        		        "configuration": {
        		          "credential-type": "X509_PROVIDED",
        		          "certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----"
        		        }
        		      }
        		    }
        		  ]
			  },
			  "no-instances-auth-method": {
	           "certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----"
			  }
	       }
	     }
	   }`
	substitutedAssignedTenantConfigurationWithLocalInstancesWithJsonpaths := `{
	     "credentials": {
	       "inboundCommunication": {
			  "auth_method_with_local_instances": {
        		  "tokenServiceUrl": "url",
        		  "clientId": "clientid",
        		  "certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----",
        		  "correlationIds": ["ASD"],
        		  "serviceInstances": [
        		    {
        		      "service": "service-test",
        		      "plan": "plan-test",
        		      "configuration": {},
        		      "serviceBinding": {
        		        "configuration": {}
        		      }
        		    },
        		    {
        		      "service": "service-test-2",
        		      "plan": "plan-test-2",
        		      "configuration": {
        		        "consumed-services": [
        		          {
        		            "first-service-instance-plan": "plan-test"
        		          }
        		        ],
        		        "xsuaa-cross-consumption": true
        		      },
        		      "serviceBinding": {
						"url": "url",
						"clientid": "clientid",
        		        "configuration": {
        		          "credential-type": "X509_PROVIDED",
        		          "certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----"
        		        }
        		      }
        		    }
        		  ]
			  },
			  "no-instances-auth-method": {
	           "certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----"
			  }
	       }
	     }
	   }`
	expectedResponseForLocalInstances := `{"state":"READY","configuration":{"credentials":{"outboundCommunication":{"auth_method_with_local_instances":{"certificate":"-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----","clientId":"clientid","correlationIds":["ASD"],"tokenServiceUrl":"url"},"no-instances-auth-method":{"certificate":"-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----"}}}}}`

	receiverTenantFormatter := `{
		"deploymentRegion": %q,
		"subaccountId": %q,
		"configuration": %s
	}`

	receiverTenantConfigurationWithServiceInstanceDetails := `{
	     "credentials": {
	       "inboundCommunication": {
			  "serviceInstances": [
	           {
	             "service": "procurement-service",
	             "plan": "apiaccess",
	             "parameters": {}
	           },
	           {
	             "service": "identity",
	             "plan": "application",
	             "xsuaa-cross-consumption": true
	           }
	         ],
			  "refering-global-instances-auth-method": {
			   "global-instance-plan": "$.credentials.inboundCommunication.serviceInstances[0].plan",
	           "certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----"
			  }
	       }
	     }
	   }`

	receiverTenantConfigurationWithServiceInstanceDetailsAndMethodWithoutInstances := `{
	    "credentials": {
	      "inboundCommunication": {
			  "serviceInstances": [
	          {
	            "service": "procurement-service",
	            "plan": "apiaccess",
	            "parameters": {}
	          },
	          {
	            "service": "identity",
	            "plan": "application",
	            "xsuaa-cross-consumption": true
	          }
	        ],
			  "no-instances-details-method": {
	           "certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----"
			  }
	      }
	    }
	  }`
	expectedResponseForGlobalInstancesWithInbound := `{"state":"READY","configuration":{"credentials":{"inboundCommunication":{"no-instances-details-method":{"certificate":"-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----"}},"outboundCommunication":{"no-instances-auth-method":{"certificate":"-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----"},"refering-global-instances-auth-method":{"certificate":"-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----"}}}}}`

	receiverTenantConfigurationWithServiceInstanceDetailsAndMethodWithoutInstancesAndReversePaths := `{
	    "credentials": {
	      "inboundCommunication": {
			  "serviceInstances": [
	          {
	            "service": "procurement-service",
	            "plan": "apiaccess",
	            "parameters": {}
	          },
	          {
	            "service": "identity",
	            "plan": "application",
	            "xsuaa-cross-consumption": true
	          }
	        ],
			  "no-instances-details-method": {
	            "certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----"
			  },
 			  "reverse-paths-method": {
	            "reverse-second-instance-plan": "$.reverse.credentials.inboundCommunication.serviceInstances[1].plan"
			  }
	      }
	    }
	  }`
	expectedResponseForGlobalInstancesWithInboundAndReverse := `{"state":"READY","configuration":{"credentials":{"inboundCommunication":{"no-instances-details-method":{"certificate":"-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----"},"reverse-paths-method":{"reverse-second-instance-plan":"application"}},"outboundCommunication":{"no-instances-auth-method":{"certificate":"-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----"},"refering-global-instances-auth-method":{"certificate":"-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----"}}}}}`

	assignedTenantFullConfiguration := `{
  "credentials": {
    "inboundCommunication": {
      "serviceInstances": [
        {
          "service": "global-service-service-test-1",
          "plan": "global-service-plan-test-1",
          "configuration": {},
          "serviceBinding": {
            "configuration": {}
          }
        },
        {
          "service": "global-service-service-test-2",
          "plan": "global-service-plan-test-2",
          "configuration": {
            "global-service-instances-plans": [
              {
                "first-service-instance-plan": "$.credentials.inboundCommunication.serviceInstances[0].plan",
                "second-service-instance-plan": "$.credentials.inboundCommunication.serviceInstances[1].plan"
              }
            ],
            "xsuaa-cross-consumption": true
          },
          "serviceBinding": {
            "url": "url",
            "clientid": "clientid",
            "configuration": {
              "credential-type": "X509_PROVIDED",
              "global-service-instances-plans": [
                {
                  "first-service-instance-plan": "$.credentials.inboundCommunication.serviceInstances[0].plan",
                  "second-service-instance-plan": "$.credentials.inboundCommunication.serviceInstances[1].plan"
                }
              ],
              "certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----"
            }
          }
        }
      ],
      "only_local_instances": {
        "first-instance-service": "$.credentials.inboundCommunication.only_local_instances.serviceInstances[0].service",
        "second-instance-binding-clientID": "$.credentials.inboundCommunication.only_local_instances.serviceInstances[1].serviceBinding.clientId",
        "complex_json_paths": "{$.credentials.inboundCommunication.serviceInstances[0].service}/complex/{$.credentials.inboundCommunication.only_local_instances.serviceInstances[0].plan}",
        "certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----",
        "correlationIds": [
          "CORR_ID"
        ],
        "serviceInstances": [
          {
            "service": "only-local-service-test",
            "plan": "only-local-plan-test",
            "configuration": {},
            "serviceBinding": {
              "configuration": {}
            }
          },
          {
            "service": "only-local-service-test-2",
            "plan": "only-local-service-test-2",
            "configuration": {
              "service-instances-services": [
                {
                  "first-only-local-service-instance-name": "$.credentials.inboundCommunication.only_local_instances.serviceInstances[0].service",
                  "second-only-local-service-instance-name": "$.credentials.inboundCommunication.only_local_instances.serviceInstances[1].service"
                }
              ],
              "xsuaa-cross-consumption": true
            },
            "serviceBinding": {
              "clientId": "client-id-test",
              "configuration": {
                "credential-type": "X509_PROVIDED",
                "certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----",
                "service-instance-plan": "$.credentials.inboundCommunication.only_local_instances.serviceInstances[1].plan"
              }
            }
          }
        ]
      },
      "without_instances": {
        "certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----",
        "correlationIds": [
          "CORR_ID_2"
        ]
      },
      "reffering_global_instances": {
        "certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----",
        "correlationIds": [
          "CORR_ID_3"
        ],
        "global-instances-plans": [
          {
            "first-global-instance-plan": "$.credentials.inboundCommunication.serviceInstances[0].plan",
            "second-global-instance-plan": "$.credentials.inboundCommunication.serviceInstances[1].plan"
          }
        ]
      },
      "local_and_reffering_global_instances": {
        "certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----",
        "correlationIds": [
          "CORR_ID_4"
        ],
        "serviceInstances": [
          {
            "name": "local_and_reffering_global_instances_name",
            "service": "local-and-reffering-global-service-test",
            "plan": "local-and-reffering-global-plan-test",
            "configuration": {},
            "serviceBinding": {
              "configuration": {}
            }
          },
          {
            "service": "local-and-reffering-global-service-test-2",
            "plan": "local-and-reffering-global-plan-test-2",
            "configuration": {
              "service-instances-services": [
                {
                  "first-only-local-service-instance-service": "$.credentials.inboundCommunication.local_and_reffering_global_instances.serviceInstances[0].service",
                  "second-only-local-service-instance-service": "$.credentials.inboundCommunication.local_and_reffering_global_instances.serviceInstances[1].service"
                }
              ],
              "xsuaa-cross-consumption": true
            },
            "serviceBinding": {
              "clientId": "client-id-test",
              "configuration": {
                "credential-type": "X509_PROVIDED",
                "certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----",
                "service-instance-plan": "$.credentials.inboundCommunication.local_and_reffering_global_instances.serviceInstances[1].plan"
              }
            }
          }
        ],
        "global-instances-services": [
          {
            "first-global-instance-service": "$.credentials.inboundCommunication.serviceInstances[0].service",
            "second-global-instance-service": "$.credentials.inboundCommunication.serviceInstances[1].service"
          }
        ]
      }
    }
  }
}`

	substitutedAssignedTenantFullConfiguration := `{
  "credentials": {
    "inboundCommunication": {
      "serviceInstances": [
        {
          "service": "global-service-service-test-1",
          "plan": "global-service-plan-test-1",
          "configuration": {},
          "serviceBinding": {
            "configuration": {}
          }
        },
        {
          "service": "global-service-service-test-2",
          "plan": "global-service-plan-test-2",
          "configuration": {
            "global-service-instances-plans": [
              {
                "first-service-instance-plan": "global-service-plan-test-1",
                "second-service-instance-plan": "global-service-plan-test-2"
              }
            ],
            "xsuaa-cross-consumption": true
          },
          "serviceBinding": {
            "url": "url",
            "clientid": "clientid",
            "configuration": {
              "credential-type": "X509_PROVIDED",
              "global-service-instances-plans": [
                {
                  "first-service-instance-plan": "global-service-plan-test-1",
                  "second-service-instance-plan": "global-service-plan-test-2"
                }
              ],
              "certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----"
            }
          }
        }
      ],
      "only_local_instances": {
        "first-instance-service": "only-local-service-test",
        "second-instance-binding-clientID": "client-id-test",
        "complex_json_paths": "global-service-service-test-1/complex/only-local-plan-test",
        "certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----",
        "correlationIds": [
          "CORR_ID"
        ],
        "serviceInstances": [
          {
            "service": "only-local-service-test",
            "plan": "only-local-plan-test",
            "configuration": {},
            "serviceBinding": {
              "configuration": {}
            }
          },
          {
            "service": "only-local-service-test-2",
            "plan": "only-local-service-test-2",
            "configuration": {
              "service-instances-services": [
                {
                  "first-only-local-service-instance-name": "only-local-service-test",
                  "second-only-local-service-instance-name": "only-local-service-test-2"
                }
              ],
              "xsuaa-cross-consumption": true
            },
            "serviceBinding": {
              "clientId": "client-id-test",
              "configuration": {
                "credential-type": "X509_PROVIDED",
                "certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----",
                "service-instance-plan": "only-local-service-test-2"
              }
            }
          }
        ]
      },
      "without_instances": {
        "certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----",
        "correlationIds": [
          "CORR_ID_2"
        ]
      },
      "reffering_global_instances": {
        "certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----",
        "correlationIds": [
          "CORR_ID_3"
        ],
        "global-instances-plans": [
          {
            "first-global-instance-plan": "global-service-plan-test-1",
            "second-global-instance-plan": "global-service-plan-test-2"
          }
        ]
      },
      "local_and_reffering_global_instances": {
        "certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----",
        "correlationIds": [
          "CORR_ID_4"
        ],
        "serviceInstances": [
          {
            "name": "local_and_reffering_global_instances_name",
            "service": "local-and-reffering-global-service-test",
            "plan": "local-and-reffering-global-plan-test",
            "configuration": {},
            "serviceBinding": {
              "configuration": {}
            }
          },
          {
            "service": "local-and-reffering-global-service-test-2",
            "plan": "local-and-reffering-global-plan-test-2",
            "configuration": {
              "service-instances-services": [
                {
                  "first-only-local-service-instance-service": "local-and-reffering-global-service-test",
                  "second-only-local-service-instance-service": "local-and-reffering-global-service-test-2"
                }
              ],
              "xsuaa-cross-consumption": true
            },
            "serviceBinding": {
              "clientId": "client-id-test",
              "configuration": {
                "credential-type": "X509_PROVIDED",
                "certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----",
                "service-instance-plan": "local-and-reffering-global-plan-test-2"
              }
            }
          }
        ],
        "global-instances-services": [
          {
            "first-global-instance-service": "global-service-service-test-1",
            "second-global-instance-service": "global-service-service-test-2"
          }
        ]
      }
    }
  }
}`

	receiverTenantFullConfiguration := `{
	  "credentials": {
        "inboundCommunication": {
          "serviceInstances": [
            {
              "service": "global-instance-service-test-1",
              "plan": "global-instance-service-plan-1",
              "parameters": {}
            },
            {
              "service": "global-instance-service-test-2",
              "plan": "global-instance-service-plan-2",
              "xsuaa-cross-consumption": true,
              "serviceBinding": {
                "url": "url"
              }
            }
          ],
          "no-instances-details": {
            "certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----"
          },
          "reffering-global-details": {
            "first-global-instance-plan": "$.credentials.inboundCommunication.serviceInstances[0].plan"
          },
          "local-details": {
            "serviceInstances": [
              {
                "service": "local-instance-service-test-1",
                "plan": "local-instance-service-plan-1",
                "parameters": {}
              }
            ]
          },
          "reverse-paths-method": {
            "reverse-second-global-instance-plan": "$.reverse.credentials.inboundCommunication.serviceInstances[1].plan",
            "only-local-instances-second-instance-binding-clientID": "$.reverse.credentials.inboundCommunication.only_local_instances.serviceInstances[1].serviceBinding.clientId"
          }
        },
        "outboundCommunication": {
          "auth1": {
            "correlationIds": [
              "CORR_ID"
            ]
          },
          "only_local_instances": {
            "certificate": "-----BEGIN CERTIFICATE----- outbound cert -----END CERTIFICATE-----"
          },
          "without_instances": {
            "certificate": "-----BEGIN CERTIFICATE----- outbound cert -----END CERTIFICATE-----",
            "random-field": "random-value"
          }
        }
      }
	}`

	expectedResponseForFullConfig := `{
  "state": "READY",
  "configuration": {
    "credentials": {
      "inboundCommunication": {
        "no-instances-details": {
          "certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----"
        },
        "reverse-paths-method": {
          "only-local-instances-second-instance-binding-clientID": "client-id-test",
          "reverse-second-global-instance-plan": "global-service-plan-test-2"
        }
      },
      "outboundCommunication": {
        "auth1": {
          "correlationIds": [
            "CORR_ID"
          ]
        },
        "local_and_reffering_global_instances": {
          "certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----",
          "correlationIds": [
            "CORR_ID_4"
          ],
          "global-instances-services": [
            {
              "first-global-instance-service": "global-service-service-test-1",
              "second-global-instance-service": "global-service-service-test-2"
            }
          ]
        },
        "only_local_instances": {
          "certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----",
          "complex_json_paths": "global-service-service-test-1/complex/only-local-plan-test",
          "correlationIds": [
            "CORR_ID"
          ],
          "first-instance-service": "only-local-service-test",
          "second-instance-binding-clientID": "client-id-test"
        },
        "reffering_global_instances": {
          "certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----",
          "correlationIds": [
            "CORR_ID_3"
          ],
          "global-instances-plans": [
            {
              "first-global-instance-plan": "global-service-plan-test-1",
              "second-global-instance-plan": "global-service-plan-test-2"
            }
          ]
        },
        "without_instances": {
          "certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----",
          "correlationIds": [
            "CORR_ID_2"
          ],
          "random-field": "random-value"
        }
      }
    }
  }
}`

	receiverTenantFullConfigurationWithoutOutboundCommunication := `{
	  "credentials": {
        "inboundCommunication": {
          "serviceInstances": [
            {
              "service": "global-instance-service-test-1",
              "plan": "global-instance-service-plan-1",
              "parameters": {}
            },
            {
              "service": "global-instance-service-test-2",
              "plan": "global-instance-service-plan-2",
              "xsuaa-cross-consumption": true,
              "serviceBinding": {
                "url": "url"
              }
            }
          ],
          "no-instances-details": {
            "certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----"
          },
          "reffering-global-details": {
            "first-global-instance-plan": "$.credentials.inboundCommunication.serviceInstances[0].plan"
          },
          "local-details": {
            "serviceInstances": [
              {
                "service": "local-instance-service-test-1",
                "plan": "local-instance-service-plan-1",
                "parameters": {}
              }
            ]
          },
          "reverse-paths-method": {
            "reverse-second-global-instance-plan": "$.reverse.credentials.inboundCommunication.serviceInstances[1].plan",
            "only-local-instances-second-instance-binding-clientID": "$.reverse.credentials.inboundCommunication.only_local_instances.serviceInstances[1].serviceBinding.clientId"
          }
        }
      }
	}`

	expectedResponseForFullConfigWithReceiverTenantWithoutOutbound := `{
  "state": "READY",
  "configuration": {
    "credentials": {
      "inboundCommunication": {
        "no-instances-details": {
          "certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----"
        },
        "reverse-paths-method": {
          "only-local-instances-second-instance-binding-clientID": "client-id-test",
          "reverse-second-global-instance-plan": "global-service-plan-test-2"
        }
      },
      "outboundCommunication": {
        "local_and_reffering_global_instances": {
          "certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----",
          "correlationIds": [
            "CORR_ID_4"
          ],
          "global-instances-services": [
            {
              "first-global-instance-service": "global-service-service-test-1",
              "second-global-instance-service": "global-service-service-test-2"
            }
          ]
        },
        "only_local_instances": {
          "certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----",
          "complex_json_paths": "global-service-service-test-1/complex/only-local-plan-test",
          "correlationIds": [
            "CORR_ID"
          ],
          "first-instance-service": "only-local-service-test",
          "second-instance-binding-clientID": "client-id-test"
        },
        "reffering_global_instances": {
          "certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----",
          "correlationIds": [
            "CORR_ID_3"
          ],
          "global-instances-plans": [
            {
              "first-global-instance-plan": "global-service-plan-test-1",
              "second-global-instance-plan": "global-service-plan-test-2"
            }
          ]
        },
        "without_instances": {
          "certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----",
          "correlationIds": [
            "CORR_ID_2"
          ]
        }
      }
    }
  }
}`

	testCases := []struct {
		name                 string
		smClientFn           func() *automock.Client
		mtlsClientFn         func() *automock.MtlsHTTPClient
		persistenceFn        func() (*persistenceautomock.DatabaseConnector, *persistenceautomock.AdvisoryLocker)
		requestBody          string
		expectedResponseCode int
	}{
		{
			name:        "Wrong json - fails on decoding",
			requestBody: `wrong json`,
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasBody("Request body contains badly-formed JSON")).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Missing config(empty json) - fails on validation",
			requestBody: emptyJSON,
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasBody("Context's Formation ID should be provided")).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Missing config(empty context, receiverTenant and assignedTenant) - fails on validation",
			requestBody: fmt.Sprintf(reqBodyFormatter, emptyJSON, emptyJSON, emptyJSON),
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasBody("while validating the request body")).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Missing formation ID in the context - fails on validation",
			requestBody: fmt.Sprintf(reqBodyFormatter, `{"operation": "assign"}`, emptyJSON, emptyJSON),
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasBody("Context's Formation ID should be provided")).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Missing operation in the context - fails on validation",
			requestBody: fmt.Sprintf(reqBodyFormatter, `{"uclFormationId": "formation-id", "operation": ""}`, emptyJSON, emptyJSON),
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasBody("Context's Operation is invalid")).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Wrong operation in the context - fails on validation",
			requestBody: fmt.Sprintf(reqBodyFormatter, `{"uclFormationId": "formation-id", "operation": "wrong-operation"}`, emptyJSON, emptyJSON),
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasBody("Context's Operation is invalid")).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Formation assignment is missing in the assignedTenant - fails on validation",
			requestBody: fmt.Sprintf(reqBodyFormatter, reqBodyContextWithAssign, emptyJSON, emptyJSON),
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasBody("Assigned Tenant Assignment ID should be provided")).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Region is missing in the receiverTenant - fails on validation",
			requestBody: fmt.Sprintf(reqBodyFormatter, reqBodyContextWithAssign, emptyJSON, fmt.Sprintf(assignedTenantFormatter, assignmentID, emptyJSON)),
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasBody("Receiver Tenant Region should be provided")).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Subaccount ID is missing in the receiverTenant - fails on validation",
			requestBody: fmt.Sprintf(reqBodyFormatter, reqBodyContextWithAssign, `{"deploymentRegion": "region"}`, fmt.Sprintf(assignedTenantFormatter, assignmentID, emptyJSON)),
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasBody("Receiver Tenant Subaccount ID should be provided")).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Operation is assign and inboundCommunication is missing in the assignedTenant configuration - fails on validation",
			requestBody: fmt.Sprintf(reqBodyFormatter, reqBodyContextWithAssign, fmt.Sprintf(receiverTenantFormatter, region, subaccount, emptyJSON), fmt.Sprintf(assignedTenantFormatter, assignmentID, emptyJSON)),
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasBody("Assigned tenant inbound communication is missing in the configuration")).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Operation is assign and receiverTenant has outboundCommunication but not in the same path as assignedTenant inboundCommunication - fails on validation",
			requestBody: fmt.Sprintf(reqBodyFormatter, reqBodyContextWithAssign, fmt.Sprintf(receiverTenantFormatter, region, subaccount, `{"credentials": {"another-field":{"credentials": {"outboundCommunication":{}}}}}`), fmt.Sprintf(assignedTenantFormatter, assignmentID, `{"credentials": {"inboundCommunication":{}}}`)),
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasBody(`Receiver tenant outbound communication should be in the same place as the assigned tenant inbound communication`)).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Operation is unassign and fails while retrieving service instances by assignment ID",
			requestBody: fmt.Sprintf(reqBodyFormatter, reqBodyContextWithUnassign, fmt.Sprintf(receiverTenantFormatter, region, subaccount, `{"credentials": {"another-field":{"credentials": {"outboundCommunication":{}}}}}`), fmt.Sprintf(assignedTenantFormatter, assignmentID, `{"credentials": {"inboundCommunication":{}}}`)),
			smClientFn: func() *automock.Client {
				client := &automock.Client{}
				client.On("RetrieveMultipleResourcesIDsByLabels", mock.Anything, region, subaccount, mock.Anything, map[string][]string{assignmentIDKey: {assignmentID}}).Return(nil, testErr).Once()
				return client
			},
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasBody(`while retrieving service instances for assignmentID`)).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			persistenceFn:        mockPersistence(assignmentID, unassignOperation),
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Operation is unassign and fails while retrieving service instances bindings by service instances IDs",
			requestBody: fmt.Sprintf(reqBodyFormatter, reqBodyContextWithUnassign, fmt.Sprintf(receiverTenantFormatter, region, subaccount, `{"credentials": {"another-field":{"credentials": {"outboundCommunication":{}}}}}`), fmt.Sprintf(assignedTenantFormatter, assignmentID, `{"credentials": {"inboundCommunication":{}}}`)),
			smClientFn: func() *automock.Client {
				client := &automock.Client{}
				client.On("RetrieveMultipleResourcesIDsByLabels", mock.Anything, region, subaccount, mock.Anything, map[string][]string{assignmentIDKey: {assignmentID}}).Return(serviceInstancesIDs, nil).Once()
				client.On("RetrieveMultipleResources", mock.Anything, region, subaccount, mock.Anything, &types.ServiceKeyMatchParameters{ServiceInstancesIDs: serviceInstancesIDs}).Return(nil, testErr).Once()
				return client
			},
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasBody(fmt.Sprintf("while retrieving service bindings for service instaces with IDs: %v", serviceInstancesIDs))).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			persistenceFn:        mockPersistence(assignmentID, unassignOperation),
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Operation is unassign and fails while deleting service instances bindings by service instances IDs",
			requestBody: fmt.Sprintf(reqBodyFormatter, reqBodyContextWithUnassign, fmt.Sprintf(receiverTenantFormatter, region, subaccount, `{"credentials": {"another-field":{"credentials": {"outboundCommunication":{}}}}}`), fmt.Sprintf(assignedTenantFormatter, assignmentID, `{"credentials": {"inboundCommunication":{}}}`)),
			smClientFn: func() *automock.Client {
				client := &automock.Client{}
				client.On("RetrieveMultipleResourcesIDsByLabels", mock.Anything, region, subaccount, mock.Anything, map[string][]string{assignmentIDKey: {assignmentID}}).Return(serviceInstancesIDs, nil).Once()
				client.On("RetrieveMultipleResources", mock.Anything, region, subaccount, mock.Anything, &types.ServiceKeyMatchParameters{ServiceInstancesIDs: serviceInstancesIDs}).Return(serviceInstancesBindingsIDs, nil).Once()
				client.On("DeleteMultipleResourcesByIDs", mock.Anything, region, subaccount, mock.Anything, serviceInstancesBindingsIDs).Return(testErr).Once()
				return client
			},
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasBody(fmt.Sprintf("while deleting service bindings with IDs: %v", serviceInstancesBindingsIDs))).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			persistenceFn:        mockPersistence(assignmentID, unassignOperation),
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Operation is unassign and fails while deleting service instances by service instances IDs",
			requestBody: fmt.Sprintf(reqBodyFormatter, reqBodyContextWithUnassign, fmt.Sprintf(receiverTenantFormatter, region, subaccount, `{"credentials": {"another-field":{"credentials": {"outboundCommunication":{}}}}}`), fmt.Sprintf(assignedTenantFormatter, assignmentID, `{"credentials": {"inboundCommunication":{}}}`)),
			smClientFn: func() *automock.Client {
				client := &automock.Client{}
				client.On("RetrieveMultipleResourcesIDsByLabels", mock.Anything, region, subaccount, mock.Anything, map[string][]string{assignmentIDKey: {assignmentID}}).Return(serviceInstancesIDs, nil).Once()
				client.On("RetrieveMultipleResources", mock.Anything, region, subaccount, mock.Anything, &types.ServiceKeyMatchParameters{ServiceInstancesIDs: serviceInstancesIDs}).Return(serviceInstancesBindingsIDs, nil).Once()
				client.On("DeleteMultipleResourcesByIDs", mock.Anything, region, subaccount, mock.Anything, serviceInstancesBindingsIDs).Return(nil).Once()
				client.On("DeleteMultipleResourcesByIDs", mock.Anything, region, subaccount, mock.Anything, serviceInstancesIDs).Return(testErr).Once()
				return client
			},
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasBody(fmt.Sprintf("while deleting service instances with IDs: %v", serviceInstancesIDs))).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			persistenceFn:        mockPersistence(assignmentID, unassignOperation),
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Success - Operation is unassign and successfully deletes instances",
			requestBody: fmt.Sprintf(reqBodyFormatter, reqBodyContextWithUnassign, fmt.Sprintf(receiverTenantFormatter, region, subaccount, `{"credentials": {"another-field":{"credentials": {"outboundCommunication":{}}}}}`), fmt.Sprintf(assignedTenantFormatter, assignmentID, `{"credentials": {"inboundCommunication":{}}}`)),
			smClientFn: func() *automock.Client {
				client := &automock.Client{}
				client.On("RetrieveMultipleResourcesIDsByLabels", mock.Anything, region, subaccount, mock.Anything, map[string][]string{assignmentIDKey: {assignmentID}}).Return(serviceInstancesIDs, nil).Once()
				client.On("RetrieveMultipleResources", mock.Anything, region, subaccount, mock.Anything, &types.ServiceKeyMatchParameters{ServiceInstancesIDs: serviceInstancesIDs}).Return(serviceInstancesBindingsIDs, nil).Once()
				client.On("DeleteMultipleResourcesByIDs", mock.Anything, region, subaccount, mock.Anything, serviceInstancesBindingsIDs).Return(nil).Once()
				client.On("DeleteMultipleResourcesByIDs", mock.Anything, region, subaccount, mock.Anything, serviceInstancesIDs).Return(nil).Once()
				return client
			},
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasJSONBody(`{"state":"READY","configuration":null}`)).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			persistenceFn:        mockPersistence(assignmentID, unassignOperation),
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Success - Operation is assign and service instances are missing. Expecting CONFIG_PENDING.",
			requestBody: fmt.Sprintf(reqBodyFormatter, reqBodyContextWithAssign, fmt.Sprintf(receiverTenantFormatter, region, subaccount, emptyJSON), fmt.Sprintf(assignedTenantFormatter, assignmentID, `{"credentials": {"inboundCommunication":{}}}`)),
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasJSONBody(`{"state":"CONFIG_PENDING","configuration":null}`)).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			persistenceFn:        mockPersistence(assignmentID, assignOperation),
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Success - Operation is assign and there are only global service instances without jsonpaths.",
			requestBody: fmt.Sprintf(reqBodyFormatter, reqBodyContextWithAssign, fmt.Sprintf(receiverTenantFormatter, region, subaccount, emptyJSON), fmt.Sprintf(assignedTenantFormatter, assignmentID, assignedTenantConfigurationWithGlobalInstancesWithoutJsonpaths)),
			smClientFn: func() *automock.Client {
				client := &automock.Client{}
				mockSMClient(client, assignedTenantConfigurationWithGlobalInstancesWithoutJsonpaths)
				return client
			},
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasJSONBody(expectedResponseForGlobalInstances)).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			persistenceFn:        mockPersistence(assignmentID, assignOperation),
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Success - Operation is assign and there are only global service instances with jsonpaths.",
			requestBody: fmt.Sprintf(reqBodyFormatter, reqBodyContextWithAssign, fmt.Sprintf(receiverTenantFormatter, region, subaccount, emptyJSON), fmt.Sprintf(assignedTenantFormatter, assignmentID, assignedTenantConfigurationWithGlobalInstancesWithJsonpaths)),
			smClientFn: func() *automock.Client {
				client := &automock.Client{}
				mockSMClient(client, substitutedAssignedTenantConfigurationWithGlobalInstancesWithJsonpaths)
				return client
			},
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasJSONBody(expectedResponseForGlobalInstances)).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			persistenceFn:        mockPersistence(assignmentID, assignOperation),
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Success - Operation is assign and there are only global service instances with jsonpaths which must be recreated.",
			requestBody: fmt.Sprintf(reqBodyFormatter, reqBodyContextWithAssign, fmt.Sprintf(receiverTenantFormatter, region, subaccount, emptyJSON), fmt.Sprintf(assignedTenantFormatter, assignmentID, assignedTenantConfigurationWithGlobalInstancesWithJsonpaths)),
			smClientFn: func() *automock.Client {
				client := &automock.Client{}
				substitutedGlobalServiceInstances := Configuration(substitutedAssignedTenantConfigurationWithGlobalInstancesWithJsonpaths).GetGlobalServiceInstances(inboundCommunicationKey).ToArray()
				client.On("RetrieveMultipleResourcesIDsByLabels", mock.Anything, region, subaccount, mock.Anything, smLabelsThatHaveAssignmentID(assignmentID)).Return(serviceInstancesIDs, nil).Once()
				// Delete All Instances and Bindings
				client.On("RetrieveMultipleResources", mock.Anything, region, subaccount, mock.Anything, &types.ServiceKeyMatchParameters{ServiceInstancesIDs: serviceInstancesIDs}).Return(serviceInstancesBindingsIDs, nil).Once()
				client.On("DeleteMultipleResourcesByIDs", mock.Anything, region, subaccount, mock.Anything, serviceInstancesBindingsIDs).Return(nil).Once()
				client.On("DeleteMultipleResourcesByIDs", mock.Anything, region, subaccount, mock.Anything, serviceInstancesIDs).Return(nil).Once()

				// First Instance
				firstGlobalServiceInstanceSubstituted := substitutedGlobalServiceInstances[0]
				client.On("RetrieveResource", mock.Anything, region, subaccount, mock.Anything, &types.ServiceOfferingMatchParameters{CatalogName: firstGlobalServiceInstanceSubstituted.GetService()}).Return(serviceOfferingIDs[0], nil).Once()
				client.On("RetrieveResource", mock.Anything, region, subaccount, mock.Anything, &types.ServicePlanMatchParameters{PlanName: firstGlobalServiceInstanceSubstituted.GetPlan(), OfferingID: serviceOfferingIDs[0]}).Return(servicePlanIDs[0], nil).Once()
				client.On("CreateResource", mock.Anything, region, subaccount, serviceInstanceReqBody(firstGlobalServiceInstanceSubstituted.GetName(), servicePlanIDs[0], assignmentID, firstGlobalServiceInstanceSubstituted.GetParameters()), mock.Anything).Return(serviceInstancesIDs[0], nil).Once()
				client.On("RetrieveRawResourceByID", mock.Anything, region, subaccount, &types.ServiceInstance{ID: serviceInstancesIDs[0]}).Return(firstGlobalServiceInstanceSubstituted.WithName(serviceInstancesNames[0]).ToJSONRawMessage(), nil).Once()

				firstGlobalServiceInstanceBindingSubstituted := firstGlobalServiceInstanceSubstituted.GetServiceBinding()
				client.On("CreateResource", mock.Anything, region, subaccount, serviceBindingReqBody(firstGlobalServiceInstanceBindingSubstituted.GetName(), serviceInstancesIDs[0], firstGlobalServiceInstanceBindingSubstituted.GetParameters()), mock.Anything).Return(serviceInstancesBindingsIDs[0], nil).Once()
				client.On("RetrieveRawResourceByID", mock.Anything, region, subaccount, &types.ServiceKey{ID: serviceInstancesBindingsIDs[0]}).Return(firstGlobalServiceInstanceBindingSubstituted.WithName(serviceInstanceBindingsNames[0]).ToJSONRawMessage(), nil).Once()

				// Second Instance
				secondGlobalServiceInstanceSubstituted := substitutedGlobalServiceInstances[1]
				client.On("RetrieveResource", mock.Anything, region, subaccount, mock.Anything, &types.ServiceOfferingMatchParameters{CatalogName: secondGlobalServiceInstanceSubstituted.GetService()}).Return(serviceOfferingIDs[1], nil).Once()
				client.On("RetrieveResource", mock.Anything, region, subaccount, mock.Anything, &types.ServicePlanMatchParameters{PlanName: secondGlobalServiceInstanceSubstituted.GetPlan(), OfferingID: serviceOfferingIDs[1]}).Return(servicePlanIDs[1], nil).Once()
				client.On("CreateResource", mock.Anything, region, subaccount, serviceInstanceReqBody(secondGlobalServiceInstanceSubstituted.GetName(), servicePlanIDs[1], assignmentID, secondGlobalServiceInstanceSubstituted.GetParameters()), mock.Anything).Return(serviceInstancesIDs[1], nil).Once()
				client.On("RetrieveRawResourceByID", mock.Anything, region, subaccount, &types.ServiceInstance{ID: serviceInstancesIDs[1]}).Return(secondGlobalServiceInstanceSubstituted.WithName(serviceInstancesNames[1]).ToJSONRawMessage(), nil).Once()

				secondGlobalServiceInstanceBindingSubstituted := secondGlobalServiceInstanceSubstituted.GetServiceBinding()
				client.On("CreateResource", mock.Anything, region, subaccount, serviceBindingReqBody(secondGlobalServiceInstanceBindingSubstituted.GetName(), serviceInstancesIDs[1], secondGlobalServiceInstanceBindingSubstituted.GetParameters()), mock.Anything).Return(serviceInstancesBindingsIDs[1], nil).Once()
				client.On("RetrieveRawResourceByID", mock.Anything, region, subaccount, &types.ServiceKey{ID: serviceInstancesBindingsIDs[1]}).Return(secondGlobalServiceInstanceBindingSubstituted.WithName(serviceInstanceBindingsNames[1]).ToJSONRawMessage(), nil).Once()

				return client
			},
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasJSONBody(expectedResponseForGlobalInstances)).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			persistenceFn:        mockPersistence(assignmentID, assignOperation),
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Success - Operation is assign and there are only global service instances and with service details in receiver tenant inbound communication - check that the inbound is deleted",
			requestBody: fmt.Sprintf(reqBodyFormatter, reqBodyContextWithAssign, fmt.Sprintf(receiverTenantFormatter, region, subaccount, receiverTenantConfigurationWithServiceInstanceDetails), fmt.Sprintf(assignedTenantFormatter, assignmentID, assignedTenantConfigurationWithGlobalInstancesWithoutJsonpaths)),
			smClientFn: func() *automock.Client {
				client := &automock.Client{}
				mockSMClient(client, assignedTenantConfigurationWithGlobalInstancesWithoutJsonpaths)
				return client
			},
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasJSONBody(expectedResponseForGlobalInstances)).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			persistenceFn:        mockPersistence(assignmentID, assignOperation),
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Success - Operation is assign and there are only global service instances and with service details in receiver tenant inbound communication - check that the inbound is left only with auth methods without service instance details",
			requestBody: fmt.Sprintf(reqBodyFormatter, reqBodyContextWithAssign, fmt.Sprintf(receiverTenantFormatter, region, subaccount, receiverTenantConfigurationWithServiceInstanceDetailsAndMethodWithoutInstances), fmt.Sprintf(assignedTenantFormatter, assignmentID, assignedTenantConfigurationWithGlobalInstancesWithoutJsonpaths)),
			smClientFn: func() *automock.Client {
				client := &automock.Client{}
				mockSMClient(client, assignedTenantConfigurationWithGlobalInstancesWithoutJsonpaths)
				return client
			},
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasJSONBody(expectedResponseForGlobalInstancesWithInbound)).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			persistenceFn:        mockPersistence(assignmentID, assignOperation),
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Success - Operation is assign and there are only global service instances and with service details in receiver tenant inbound communication - check that the inbound is left only with auth methods without service instance details",
			requestBody: fmt.Sprintf(reqBodyFormatter, reqBodyContextWithAssign, fmt.Sprintf(receiverTenantFormatter, region, subaccount, receiverTenantConfigurationWithServiceInstanceDetailsAndMethodWithoutInstancesAndReversePaths), fmt.Sprintf(assignedTenantFormatter, assignmentID, assignedTenantConfigurationWithGlobalInstancesWithoutJsonpaths)),
			smClientFn: func() *automock.Client {
				client := &automock.Client{}
				mockSMClient(client, assignedTenantConfigurationWithGlobalInstancesWithoutJsonpaths)
				return client
			},
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasJSONBody(expectedResponseForGlobalInstancesWithInboundAndReverse)).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			persistenceFn:        mockPersistence(assignmentID, assignOperation),
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Success - Operation is assign and there are only local service instances with jsonpaths.",
			requestBody: fmt.Sprintf(reqBodyFormatter, reqBodyContextWithAssign, fmt.Sprintf(receiverTenantFormatter, region, subaccount, emptyJSON), fmt.Sprintf(assignedTenantFormatter, assignmentID, assignedTenantConfigurationWithLocalInstancesWithJsonpaths)),
			smClientFn: func() *automock.Client {
				client := &automock.Client{}
				mockSMClient(client, substitutedAssignedTenantConfigurationWithLocalInstancesWithJsonpaths)
				return client
			},
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasJSONBody(expectedResponseForLocalInstances)).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			persistenceFn:        mockPersistence(assignmentID, assignOperation),
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Success - Operation is assign with full config",
			requestBody: fmt.Sprintf(reqBodyFormatter, reqBodyContextWithAssign, fmt.Sprintf(receiverTenantFormatter, region, subaccount, receiverTenantFullConfiguration), fmt.Sprintf(assignedTenantFormatter, assignmentID, assignedTenantFullConfiguration)),
			smClientFn: func() *automock.Client {
				client := &automock.Client{}
				mockSMClient(client, substitutedAssignedTenantFullConfiguration)
				return client
			},
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasJSONBody(expectedResponseForFullConfig)).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			persistenceFn:        mockPersistence(assignmentID, assignOperation),
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Success - Operation is assign with full config but receiver tenant outboundCommunication missing",
			requestBody: fmt.Sprintf(reqBodyFormatter, reqBodyContextWithAssign, fmt.Sprintf(receiverTenantFormatter, region, subaccount, receiverTenantFullConfigurationWithoutOutboundCommunication), fmt.Sprintf(assignedTenantFormatter, assignmentID, assignedTenantFullConfiguration)),
			smClientFn: func() *automock.Client {
				client := &automock.Client{}
				mockSMClient(client, substitutedAssignedTenantFullConfiguration)
				return client
			},
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasJSONBody(expectedResponseForFullConfigWithReceiverTenantWithoutOutbound)).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			persistenceFn:        mockPersistence(assignmentID, assignOperation),
			expectedResponseCode: http.StatusAccepted,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			//GIVEN
			smClient := &automock.Client{}
			if testCase.smClientFn != nil {
				smClient = testCase.smClientFn()
			}
			mtlsClient := &automock.MtlsHTTPClient{}
			if testCase.mtlsClientFn != nil {
				mtlsClient = testCase.mtlsClientFn()
			}
			dbConnector := &persistenceautomock.DatabaseConnector{}
			advisoryLocker := &persistenceautomock.AdvisoryLocker{}
			if testCase.persistenceFn != nil {
				dbConnector, advisoryLocker = testCase.persistenceFn()
			}
			defer mock.AssertExpectationsForObjects(t, smClient, dbConnector, advisoryLocker)

			req, err := http.NewRequest(http.MethodPost, url+apiPath, bytes.NewBuffer([]byte(testCase.requestBody)))
			require.NoError(t, err)
			req.Header.Set("Location", statusUrl)

			h := handler.NewHandler(smClient, mtlsClient, dbConnector)
			recorder := httptest.NewRecorder()

			//WHEN
			h.HandlerFunc(recorder, req)
			resp := recorder.Result()

			body, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)

			require.Equal(t, testCase.expectedResponseCode, resp.StatusCode, string(body))
			require.Eventually(t, func() bool {
				return mtlsClient.AssertExpectations(t)
			}, time.Second*15, 50*time.Millisecond)
		})
	}
}

func requestThatHasBody(expectedBody string) interface{} {
	return mock.MatchedBy(func(actualReq *http.Request) bool {
		bytes, err := io.ReadAll(actualReq.Body)
		if err != nil {
			return false
		}
		fmt.Printf("Expected Body %q\n", string(bytes))
		return strings.Contains(string(bytes), expectedBody)
	})
}

func requestThatHasJSONBody(expectedBody string) interface{} {
	return mock.MatchedBy(func(actualReq *http.Request) bool {
		bytes, err := io.ReadAll(actualReq.Body)
		if err != nil {
			return false
		}
		fmt.Printf("Expected Body %q\n", string(bytes))
		var expectedJSONAsInterface, actualJSONAsInterface interface{}

		if err := json.Unmarshal([]byte(expectedBody), &expectedJSONAsInterface); err != nil {
			return false
		}

		if err := json.Unmarshal(bytes, &actualJSONAsInterface); err != nil {
			return false
		}

		return reflect.DeepEqual(actualJSONAsInterface, expectedJSONAsInterface)
	})
}

func fixHTTPResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func smLabelsThatHaveAssignmentID(expectedAssignmentID string) interface{} {
	return mock.MatchedBy(func(actualMap map[string][]string) bool {
		actualLabel, ok := actualMap[assignmentIDKey]
		return ok && len(actualLabel) == 1 && actualLabel[0] == expectedAssignmentID
	})
}

func serviceInstanceReqBody(name, planID, assignmentID, parameters string) interface{} {
	return mock.MatchedBy(func(actualReqBody *types.ServiceInstanceReqBody) bool {
		actualLabel, ok := actualReqBody.Labels[assignmentIDKey]
		return (name == "" || name == actualReqBody.Name) &&
			planID == actualReqBody.ServicePlanID &&
			parameters == string(actualReqBody.Parameters) &&
			ok && len(actualLabel) == 1 && assignmentID == actualLabel[0]
	})
}

func serviceBindingReqBody(name, serviceInstanceID, parameters string) interface{} {
	return mock.MatchedBy(func(actualReqBody *types.ServiceKeyReqBody) bool {
		return (name == "" || name == actualReqBody.Name) &&
			parameters == string(actualReqBody.Parameters) &&
			serviceInstanceID == actualReqBody.ServiceKeyID
	})
}

type Configuration string
type ServiceInstances string
type ServiceInstance string
type ServiceBinding string

func (c Configuration) GetCommunication(typeCommunication string) string {
	return gjson.Get(string(c), tenantmapping.FindKeyPath(gjson.Parse(string(c)).Value(), typeCommunication)).String()
}

func (c Configuration) GetGlobalServiceInstances(typeCommunication string) ServiceInstances {
	return ServiceInstances(gjson.Get(string(c), fmt.Sprintf("%s.%s", tenantmapping.FindKeyPath(gjson.Parse(string(c)).Value(), typeCommunication), serviceInstancesKey)).String())
}

func (c Configuration) GetLocalServiceInstances(typeCommunication, authMethod string) ServiceInstances {
	return ServiceInstances(gjson.Get(string(c), fmt.Sprintf("%s.%s.%s", tenantmapping.FindKeyPath(gjson.Parse(string(c)).Value(), typeCommunication), authMethod, serviceInstancesKey)).String())
}

func (c Configuration) GetAuthMethodsWithLocalServiceInstances(typeCommunication string) []string {
	var res []string

	gjson.Get(string(c), tenantmapping.FindKeyPath(gjson.Parse(string(c)).Value(), typeCommunication)).ForEach(func(key, value gjson.Result) bool {
		if gjson.Get(value.Raw, serviceInstancesKey).Exists() {
			res = append(res, key.String())
		}

		return true
	})

	return res
}

func (sis ServiceInstances) ToArray() []ServiceInstance {
	arr := gjson.Parse(string(sis)).Array()
	result := make([]ServiceInstance, 0, len(arr))
	for _, el := range arr {
		result = append(result, ServiceInstance(el.String()))
	}
	return result
}

func (sis ServiceInstances) ToString() string {
	return string(sis)
}

func (si ServiceInstance) GetService() string {
	return gjson.Get(string(si), serviceInstanceServiceKey).String()
}

func (si ServiceInstance) GetPlan() string {
	return gjson.Get(string(si), serviceInstancePlanKey).String()
}

func (si ServiceInstance) GetName() string {
	return gjson.Get(string(si), nameKey).String()
}

func (si ServiceInstance) GetParameters() string {
	return gjson.Get(string(si), configurationKey).String()
}

func (si ServiceInstance) GetServiceBinding() ServiceBinding {
	return ServiceBinding(gjson.Get(string(si), serviceBindingKey).String())
}

func (si ServiceInstance) WithName(name string) ServiceInstance {
	instanceWithName, _ := sjson.Set(string(si), nameKey, name)
	return ServiceInstance(instanceWithName)
}

func (si ServiceInstance) ToJSONRawMessage() json.RawMessage {
	return []byte(si)
}

func (sb ServiceBinding) GetParameters() string {
	return gjson.Get(string(sb), configurationKey).String()
}

func (sb ServiceBinding) GetName() string {
	return gjson.Get(string(sb), nameKey).String()
}

func (sb ServiceBinding) WithName(name string) ServiceBinding {
	bindingWithName, _ := sjson.Set(string(sb), nameKey, name)
	return ServiceBinding(bindingWithName)
}

func (sb ServiceBinding) ToJSONRawMessage() json.RawMessage {
	return []byte(sb)
}

func mockSMClient(client *automock.Client, assignedTenantConfiguration string) {
	config := Configuration(assignedTenantConfiguration)

	// Global Instances
	substitutedGlobalServiceInstances := config.GetGlobalServiceInstances(inboundCommunicationKey).ToArray()

	if len(substitutedGlobalServiceInstances) != 0 {
		client.On("RetrieveMultipleResourcesIDsByLabels", mock.Anything, region, subaccount, mock.Anything, smLabelsThatHaveAssignmentID(assignmentID)).Return(nil, nil).Once()
	}

	for i, globalInstance := range substitutedGlobalServiceInstances {
		currentServiceOfferingID := fmt.Sprintf("%s-%d", serviceOfferingID, i)
		currentServicePlanID := fmt.Sprintf("%s-%d", servicePlanID, i)
		currentServiceInstanceID := fmt.Sprintf("%s-%d", serviceInstanceID, i)
		currentServiceBindingID := fmt.Sprintf("%s-%d", serviceInstanceBindingID, i)
		currentServiceInstanceBindingName := fmt.Sprintf("%s-%d", serviceInstanceBindingName, i)
		currentServiceInstanceName := fmt.Sprintf("%s-%d", serviceInstanceName, i)

		// Service Instance
		client.On("RetrieveResource", mock.Anything, region, subaccount, mock.Anything, &types.ServiceOfferingMatchParameters{CatalogName: globalInstance.GetService()}).Return(currentServiceOfferingID, nil).Once()
		client.On("RetrieveResource", mock.Anything, region, subaccount, mock.Anything, &types.ServicePlanMatchParameters{PlanName: globalInstance.GetPlan(), OfferingID: currentServiceOfferingID}).Return(currentServicePlanID, nil).Once()
		client.On("CreateResource", mock.Anything, region, subaccount, serviceInstanceReqBody(globalInstance.GetName(), currentServicePlanID, assignmentID, globalInstance.GetParameters()), mock.Anything).Return(currentServiceInstanceID, nil).Once()
		client.On("RetrieveRawResourceByID", mock.Anything, region, subaccount, &types.ServiceInstance{ID: currentServiceInstanceID}).Return(globalInstance.WithName(currentServiceInstanceName).ToJSONRawMessage(), nil).Once()

		// Service Binding
		globalInstanceBinding := globalInstance.GetServiceBinding()
		client.On("CreateResource", mock.Anything, region, subaccount, serviceBindingReqBody(globalInstanceBinding.GetName(), currentServiceInstanceID, globalInstanceBinding.GetParameters()), mock.Anything).Return(currentServiceBindingID, nil).Once()
		client.On("RetrieveRawResourceByID", mock.Anything, region, subaccount, &types.ServiceKey{ID: currentServiceBindingID}).Return(globalInstanceBinding.WithName(currentServiceInstanceBindingName).ToJSONRawMessage(), nil).Once()
	}

	// Local Instances
	authMethodsWithLocalInstances := config.GetAuthMethodsWithLocalServiceInstances(inboundCommunicationKey)

	for _, auth := range authMethodsWithLocalInstances {
		localInstances := config.GetLocalServiceInstances(inboundCommunicationKey, auth).ToArray()

		client.On("RetrieveMultipleResourcesIDsByLabels", mock.Anything, region, subaccount, mock.Anything, smLabelsThatHaveAssignmentID(assignmentID)).Return(nil, nil).Once()

		for i, localInstance := range localInstances {
			currentServiceOfferingID := fmt.Sprintf("%s-local-%d", serviceOfferingID, i)
			currentServicePlanID := fmt.Sprintf("%s-local-%d", servicePlanID, i)
			currentServiceInstanceID := fmt.Sprintf("%s-local-%d", serviceInstanceID, i)
			currentServiceBindingID := fmt.Sprintf("%s-local-%d", serviceInstanceBindingID, i)
			currentServiceInstanceBindingName := fmt.Sprintf("%s-local-%d", serviceInstanceBindingName, i)
			currentServiceInstanceName := fmt.Sprintf("%s-local-%d", serviceInstanceName, i)

			// Service Instance
			client.On("RetrieveResource", mock.Anything, region, subaccount, mock.Anything, &types.ServiceOfferingMatchParameters{CatalogName: localInstance.GetService()}).Return(currentServiceOfferingID, nil).Once()
			client.On("RetrieveResource", mock.Anything, region, subaccount, mock.Anything, &types.ServicePlanMatchParameters{PlanName: localInstance.GetPlan(), OfferingID: currentServiceOfferingID}).Return(currentServicePlanID, nil).Once()
			client.On("CreateResource", mock.Anything, region, subaccount, serviceInstanceReqBody(localInstance.GetName(), currentServicePlanID, assignmentID, localInstance.GetParameters()), mock.Anything).Return(currentServiceInstanceID, nil).Once()
			client.On("RetrieveRawResourceByID", mock.Anything, region, subaccount, &types.ServiceInstance{ID: currentServiceInstanceID}).Return(localInstance.WithName(currentServiceInstanceName).ToJSONRawMessage(), nil).Once()

			// Service Binding
			localInstanceBinding := localInstance.GetServiceBinding()
			client.On("CreateResource", mock.Anything, region, subaccount, serviceBindingReqBody(localInstanceBinding.GetName(), currentServiceInstanceID, localInstanceBinding.GetParameters()), mock.Anything).Return(currentServiceBindingID, nil).Once()
			client.On("RetrieveRawResourceByID", mock.Anything, region, subaccount, &types.ServiceKey{ID: currentServiceBindingID}).Return(localInstanceBinding.WithName(currentServiceInstanceBindingName).ToJSONRawMessage(), nil).Once()
		}
	}
}

func mockPersistence(assignmentID, operation string) func() (*persistenceautomock.DatabaseConnector, *persistenceautomock.AdvisoryLocker) {
	return func() (*persistenceautomock.DatabaseConnector, *persistenceautomock.AdvisoryLocker) {
		connection := &persistenceautomock.Connection{}

		dbConnector := &persistenceautomock.DatabaseConnector{}
		dbConnector.On("GetConnection", mock.Anything).Return(connection, nil).Once()

		locker := &persistenceautomock.AdvisoryLocker{}
		connection.On("GetAdvisoryLocker").Return(locker).Once()
		connection.On("Close").Return(nil).Once()

		locker.On("TryLock", mock.Anything, assignmentID+operation).Return(true, nil).Once()
		locker.On("Unlock", mock.Anything, assignmentID+operation).Return(nil).Once()
		locker.On("Lock", mock.Anything, assignmentID).Return(nil).Once()
		locker.On("Unlock", mock.Anything, assignmentID).Return(nil).Once()

		return dbConnector, locker
	}
}
