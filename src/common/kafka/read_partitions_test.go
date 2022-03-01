//go:build !tests

/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package kafka

const realCreds string = `{
}`

//func TestReadPartitions(t *testing.T) {
//
//	d := &kg.Dialer{
//		//SASLMechanism: plain.Mechanism{"token", "REPLACE-ME-WITH-PASSWORD"},
//		TLS: &tls.Config{
//			MaxVersion: tls.VersionTLS12,
//		},
//	}
//
//	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
//	defer cancel()
//
//	//conn, err := d.DialLeader (ctx, "tcp", "broker-0-g1xtlwh2v9x9rtcp.kafka.svc01.us-south.eventstreams.cloud.ibm.com:9093", "connect-offsets", 0)
//	conn, err := d.DialContext(ctx, "tcp", "broker-2-g1xtlwh2v9x9rtcp.kafka.svc01.us-south.eventstreams.cloud.ibm.com:9093")
//
//	if err != nil {
//		t.Error(err)
//		conn.Close()
//	}
//
//	defer conn.Close()
//	parts, err := conn.ReadPartitions()
//	if err != nil {
//		t.Error(err)
//	}
//
//	if len(parts) == 0 {
//		t.Errorf("no partitions were returned")
//	} else {
//		fmt.Printf("partitions Count: %v \n", len(parts))
//		part := parts[0]
//		fmt.Printf("Topic: %v", part.Topic+"| firstPartition: "+strconv.Itoa(part.ID)+" \n")
//		fmt.Printf("partitions: %v \n", parts)
//
//	}
//
//	conn.Close()
//}
