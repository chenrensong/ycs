// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

using System.Collections.Generic;

namespace Ycs.Contracts
{
    public interface ISnapshot
    {
        IDeleteSet DeleteSet { get; }
        IDictionary<long, long> StateVector { get; }
        byte[] EncodeSnapshotV2();
        IYDoc RestoreDocument(IYDoc originDoc, YDocOptions opts = null);
    }
}