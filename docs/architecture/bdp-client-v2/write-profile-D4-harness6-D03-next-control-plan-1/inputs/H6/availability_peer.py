"""CLOSED zero-product peer. Executed only after actual bootstrap ACK; no subject import."""
import time
FIRST=time.monotonic()
import json,os,signal,sys

def main():
 if len(sys.argv)!=4 or sys.argv[1]not in {'term','kill'}:return 1
 role,nonce,origin=sys.argv[1],sys.argv[2],float(sys.argv[3])
 mask=signal.pthread_sigmask(signal.SIG_BLOCK,set())
 if signal.SIGTERM in mask or signal.SIGINT in mask:return 1
 observed=[]
 signal.signal(signal.SIGTERM,lambda sig,frame:observed.append(int(sig)))
 if role=='kill':os.kill(os.getpid(),signal.SIGKILL);return 1
 os.kill(os.getpid(),signal.SIGTERM)
 if observed!=[int(signal.SIGTERM)]:return 1
 value={'event':'availability-peer','role':role,'nonce':nonce,'pid':os.getpid(),'parent':os.getppid(),'pgid':os.getpgrp(),'firstMonotonic':FIRST,'lastMonotonic':time.monotonic(),'nativeOrigin':origin,'termIntUnblocked':True,'observedSignals':observed,'productCalls':0}
 os.set_blocking(1,False);raw=(json.dumps(value,separators=(',',':'),allow_nan=False)+'\n').encode()
 if len(raw)>4096 or os.write(1,raw)!=len(raw):return 1
 return 0
if __name__=='__main__':raise SystemExit(main())
