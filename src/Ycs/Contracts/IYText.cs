// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

using System;
using System.Collections.Generic;

namespace Ycs.Contracts
{
    public interface IYText
    {
        void ApplyDelta(IList<Delta> delta, bool sanitize = true);
        void CallObserver(ITransaction transaction, ISet<string> parentSubs);
        IYText Clone();
        void Delete(int index, int length);
        void Format(int index, int length, IDictionary<string, object> attributes);
        object GetAttribute(string name);
        IEnumerable<KeyValuePair<string, object>> GetAttributes();
        void Insert(int index, string text, IDictionary<string, object> attributes = null);
        void InsertEmbed(int index, object embed, IDictionary<string, object> attributes = null);
        void Integrate(IYDoc doc, IStructItem item);
        IAbstractType InternalClone();
        void RemoveAttribute(string name);
        void SetAttribute(string name, object value);
        IList<Delta> ToDelta(ISnapshot snapshot = null, ISnapshot prevSnapshot = null, Func<YTextChangeType, StructID, YTextChangeAttributes> computeYChange = null);
        string ToString();
        void Write(IUpdateEncoder encoder);
    }
    public class YTextChangeAttributes
    {
        public YTextChangeType Type { get; set; }
        public int User { get; set; }
        public YTextChangeType State { get; set; }
    }
    public enum YTextChangeType
    {
        Added,
        Removed
    }
}